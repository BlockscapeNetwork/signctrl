package privval

import (
	"io"
	"log"
	"net"
	"os"
	"syscall"

	"github.com/BlockscapeNetwork/signctrl/config"
	"github.com/BlockscapeNetwork/signctrl/types"
	"github.com/hashicorp/logutils"
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
type SCFilePV struct {
	types.BaseSignCtrled

	CurrentHeight int64
	Logger        *log.Logger
	Config        *config.Config
	TMFilePV      *tm_privval.FilePV
}

// KeyFilePath returns the absolute path to the priv_validator_key.json file.
func KeyFilePath(cfgDir string) string {
	return cfgDir + "/" + KeyFile
}

// StateFilePath returns the absolute path to the priv_validator_state.json file.
func StateFilePath(cfgDir string) string {
	return cfgDir + "/" + StateFile
}

// NewSCFilePV creates a new instance of SCFilePV.
func NewSCFilePV(logger *log.Logger, cfg *config.Config, tmpv *tm_privval.FilePV) *SCFilePV {
	filter := &logutils.LevelFilter{
		Levels:   config.LogLevels,
		MinLevel: logutils.LogLevel(cfg.Init.LogLevel),
		Writer:   os.Stderr,
	}
	logger.SetOutput(filter)

	pv := &SCFilePV{
		Logger:        logger,
		Config:        cfg,
		CurrentHeight: 1, // Start on genesis height
		TMFilePV:      tmpv,
	}
	pv.BaseSignCtrled = *types.NewBaseSignCtrled(
		logger,
		pv.Config.Init.Threshold,
		pv.Config.Init.Rank,
		pv,
	)

	return pv
}

// run runs the main loop of SignCTRL. It handls incoming messages from the validator.
func run(conn net.Conn, quitCh chan os.Signal, pv *SCFilePV) {
	for {
		var msg tm_privvalproto.Message
		r := tm_protoio.NewDelimitedReader(conn, maxRemoteSignerMsgSize)
		if _, err := r.ReadMsg(&msg); err != nil {
			if err == io.EOF {
				// Prevent the logs from being spammed with EOF errors
				continue
			}
			pv.Logger.Printf("[ERR] signctrl: couldn't read message: %v\n", err)
		}

		w := tm_protoio.NewDelimitedWriter(conn)
		resp, err := HandleRequest(&msg, pv)

		if _, err := w.WriteMsg(resp); err != nil {
			pv.Logger.Printf("[ERR] signctrl: couldn't write message: %v\n", err)
		}

		if err != nil {
			pv.Logger.Printf("[ERR] signctrl: couldn't handle request: %v\n", err)
			if err == types.ErrMustShutdown {
				pv.Logger.Println("[INFO] signctrl: Terminating SignCTRL...")
				r.Close()
				w.Close()
				quitCh <- syscall.SIGINT
				close(quitCh)
				return
			}
		}
	}
}

// Start starts the main loop in a separate goroutine and returns an os.Signal channel
// that is used to terminate it.
func (pv *SCFilePV) Start(conn net.Conn) <-chan os.Signal {
	pv.Logger.Printf("[INFO] signctrl: Starting SignCTRL... (rank: %v)", pv.GetRank())
	quitCh := make(chan os.Signal, 1)
	go run(conn, quitCh, pv)
	return quitCh
}
