package cmd

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/hashicorp/logutils"
	"github.com/tendermint/tendermint/libs/protoio"

	"github.com/BlockscapeNetwork/pairmint/config"
	"github.com/BlockscapeNetwork/pairmint/connection"
	"github.com/BlockscapeNetwork/pairmint/privval"
	"github.com/BlockscapeNetwork/pairmint/utils"

	tmprivval "github.com/tendermint/tendermint/privval"

	"github.com/spf13/cobra"
)

var (
	// The start command starts the pairmint application.
	startCmd = &cobra.Command{
		Use:   "start",
		Short: "Start the pairmint daemon",
		Long:  "",
		Run: func(cmd *cobra.Command, args []string) {
			// Initialize configuration directory.
			configDir := config.GetDir()

			// Configure the PrivValidator.
			pv := privval.NewPairmintFilePV()
			pv.Logger = log.New(os.Stderr, "", 0)
			if err := pv.Config.Load(); err != nil {
				pv.Logger.Printf("[ERR] pairmint: error while loading configuration: %v\n", err)
				os.Exit(int(syscall.SIGHUP))
			}
			pv.FilePV = tmprivval.LoadOrGenFilePV(pv.Config.FilePV.KeyFilePath, pv.Config.FilePV.StateFilePath)

			pv.Logger.Printf("[INFO] pairmint: Validator node is ranked #%v\n", pv.Config.Init.Rank)

			// Configure minimum log level for logger.
			pv.Logger.SetOutput(&logutils.LevelFilter{
				Levels:   []logutils.LogLevel{"DEBUG", "INFO", "WARN", "ERR"},
				MinLevel: logutils.LogLevel(pv.Config.Init.LogLevel),
				Writer:   os.Stderr,
			})

			// Load the keypair from the pm-identity.key file. The private key is necessary for
			// establishing a secret connection to Tendermint.
			priv, _, err := utils.LoadKeypair(configDir + "/pm-identity.key")
			if err != nil {
				pv.Logger.Printf("[ERR] pairmint: error while loading keypair: %v\n", err)
				os.Exit(int(syscall.SIGHUP))
			}

			// Establish a secret connection to the Tendermint validator.
			rwc := connection.NewReadWriteConn()
			rwc.SecretConn, err = connection.RetrySecretDial("tcp", pv.Config.Init.ValidatorListenAddr, priv, pv.Logger)
			if err != nil {
				pv.Logger.Printf("[ERR] pairmint: error while establishing secret connection: %v\n", err)
				os.Exit(int(syscall.SIGHUP))
			}
			defer rwc.SecretConn.Close()

			rwc.Reader = protoio.NewDelimitedReader(rwc.SecretConn, 64<<10)
			rwc.Writer = protoio.NewDelimitedWriter(rwc.SecretConn)

			pubkey, err := pv.GetPubKey()
			if err != nil {
				pv.Logger.Printf("[ERR] pairmint: couldn't get privval pubkey: %v\n", err)
				os.Exit(int(syscall.SIGHUP))
			}

			sigCh := make(chan os.Signal, 1)
			signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
			defer close(sigCh)

			// Run the routine for reading and writing messages.
			pv.Run(rwc, pubkey, sigCh)
		},
	}
)

func init() {
	// Add start to the root command.
	rootCmd.AddCommand(startCmd)
}
