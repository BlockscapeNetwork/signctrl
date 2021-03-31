package privval

import (
	"context"
	"io"
	"net"
	"net/http"
	"path/filepath"
	"time"

	"github.com/BlockscapeNetwork/signctrl/config"
	"github.com/BlockscapeNetwork/signctrl/connection"
	"github.com/BlockscapeNetwork/signctrl/types"
	tm_protoio "github.com/tendermint/tendermint/libs/protoio"
	tm_privval "github.com/tendermint/tendermint/privval"
	tm_privvalproto "github.com/tendermint/tendermint/proto/tendermint/privval"
)

const (
	// KeyFile is Tendermint's default file name for the private validator's keys.
	KeyFile = "priv_validator_key.json"

	// StateFile is Tendermint's default file name for the private validator's state.
	StateFile = "priv_validator_state.json"

	// maxRemoteSignerMsgSize determines the maximum size in bytes for the delimited
	// reader.
	maxRemoteSignerMsgSize = 1024 * 10
)

// SCFilePV must implement the SignCtrled interface.
var _ types.SignCtrled = new(SCFilePV)

// SCFilePV is a wrapper for tm_privval.FilePV.
// Implements the SignCtrled interface by embedding BaseSignCtrled.
// Implements the Service interface by embedding BaseService.
type SCFilePV struct {
	types.BaseService
	types.BaseSignCtrled

	Logger     *types.SyncLogger
	Config     config.Config
	State      *config.State
	TMFilePV   tm_privval.FilePV
	SecretConn net.Conn
	HTTP       *http.Server
	Gauges     types.Gauges
}

// KeyFilePath returns the absolute path to the priv_validator_key.json file.
func KeyFilePath(cfgDir string) string {
	return filepath.Join(cfgDir, KeyFile)
}

// StateFilePath returns the absolute path to the priv_validator_state.json file.
func StateFilePath(cfgDir string) string {
	return filepath.Join(cfgDir, StateFile)
}

// NewSCFilePV creates a new instance of SCFilePV.
func NewSCFilePV(logger *types.SyncLogger, cfg config.Config, state *config.State, tmpv tm_privval.FilePV, http *http.Server) *SCFilePV {
	pv := &SCFilePV{
		Logger:   logger,
		Config:   cfg,
		State:    state,
		TMFilePV: tmpv,
		HTTP:     http,
	}
	pv.BaseService = *types.NewBaseService(
		logger,
		"SignCTRL",
		pv,
	)
	pv.BaseSignCtrled = *types.NewBaseSignCtrled(
		logger,
		pv.Config.Base.Threshold,
		pv.Config.Base.StartRank,
		pv,
	)
	pv.Gauges = types.RegisterGauges()

	return pv
}

// run runs the main loop of SignCTRL. It handles incoming messages from the validator.
// In order to stop the goroutine, Stop() can be called outside of run(). The goroutine
// returns on its own once SignCTRL is forced to shut down.
func (pv *SCFilePV) run() {
	retryDialTimeout := config.GetRetryDialTime(pv.Config.Base.RetryDialAfter)
	timeout := time.NewTimer(retryDialTimeout)
	ctx, cancel := context.WithCancel(context.Background())

	for {
		select {
		case <-pv.Quit():
			pv.Logger.Debug("Terminating run goroutine: service stopped")
			cancel()
			// Note: Don't use pv.Stop() in here, as it closes the pv.Quit() channel.
			return

		case <-timeout.C:
			pv.Logger.Info("Lost connection to the validator... (no message for %v)\n", retryDialTimeout.String())

			// Lock the counter for missed blocks in a row again.
			pv.LockCounter()

			// Close the connection and establish a new one.
			if err := pv.SecretConn.Close(); err != nil {
				pv.Logger.Error("%v", err)
			}

			var err error
			if pv.SecretConn, err = connection.RetryDial(
				config.Dir(),
				pv.Config.Base.ValidatorListenAddress,
				pv.Logger,
			); err != nil {
				pv.Logger.Error("couldn't dial validator: %v\n", err)
				cancel()
				// Note: Don't use pv.Stop() in here, as RetryDial can only be stopped via SIGINT/SIGTERM.
				return
			}

		default:
			var msg tm_privvalproto.Message
			r := tm_protoio.NewDelimitedReader(pv.SecretConn, maxRemoteSignerMsgSize)
			if _, err := r.ReadMsg(&msg); err != nil {
				if err != io.EOF {
					pv.Logger.Error("couldn't read message: %v\n", err)
				}
				continue
			}

			timeout.Reset(retryDialTimeout)
			cancel()

			ctx, cancel = context.WithCancel(context.Background())
			resp, err := HandleRequest(ctx, &msg, pv)
			w := tm_protoio.NewDelimitedWriter(pv.SecretConn)
			if _, err := w.WriteMsg(resp); err != nil {
				pv.Logger.Error("couldn't write message: %v\n", err)
			}
			if err != nil {
				pv.Logger.Error("couldn't handle request: %v\n", err)
				if err == types.ErrMustShutdown || err == ErrRankObsolete {
					pv.Logger.Debug("Terminating run goroutine: %v\n", err)
					cancel()
					if err := pv.Stop(); err != nil {
						pv.Logger.Error("%v", err)
					}
					if err := pv.SecretConn.Close(); err != nil {
						pv.Logger.Error("%v", err)
					}
					return
				}
			}
		}
	}
}

// OnStart starts the main loop of the SignCtrled PrivValidator.
// Implements the Service interface.
func (pv *SCFilePV) OnStart() (err error) {
	pv.Logger.Info("Starting SignCTRL on rank %v...\n", pv.GetRank())

	// Start http server.
	if err := pv.StartHTTPServer(); err != nil {
		return err
	}

	// Dial the validator.
	if pv.SecretConn, err = connection.RetryDial(
		config.Dir(),
		pv.Config.Base.ValidatorListenAddress,
		pv.Logger,
	); err != nil {
		return err
	}

	// Run the main loop.
	go pv.run()

	return nil
}

// OnStop terminates the main loop of the SignCtrled PrivValidator.
// Implements the Service interface.
func (pv *SCFilePV) OnStop() error {
	pv.Logger.Info("Stopping SignCTRL on rank %v...\n", pv.GetRank())

	// Close the http server.
	pv.Logger.Info("Stopping the HTTP server...")
	pv.HTTP.Close()

	// Save rank to last_rank.json file if the shutdown was not self-induced.
	pv.State.LastRank = pv.GetRank()
	if err := pv.State.Save(config.Dir()); err != nil {
		pv.Logger.Error("couldn't save state to %v: %v\n", config.StateFile, err)
		return err
	}

	return nil
}

// OnMissedTooMany sets the prometheus gauge for the validator's counter for missed
// blocks in a row.
// Implements the SignCtrled interface.
func (pv *SCFilePV) OnMissedTooMany() {
	pv.Logger.Debug("Setting signctrl_missed_blocks_in_a_row gauge to %v\n", pv.GetMissedInARow())
	pv.Gauges.MissedInARowGauge.Set(float64(pv.GetMissedInARow()))
}

// OnPromote sets the prometheus gauge for the validator's rank.
// Implements the SignCtrled interface.
func (pv *SCFilePV) OnPromote() {
	pv.Logger.Debug("Setting signctrl_rank gauge to %v\n", pv.GetRank())
	pv.Gauges.RankGauge.Set(float64(pv.GetRank()))
}
