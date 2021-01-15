package cmd

import (
	"log"
	"os"
	"os/signal"

	"github.com/hashicorp/logutils"
	"github.com/tendermint/tendermint/libs/protoio"

	"github.com/BlockscapeNetwork/pairmint/config"
	"github.com/BlockscapeNetwork/pairmint/connection"
	"github.com/BlockscapeNetwork/pairmint/privval"
	"github.com/BlockscapeNetwork/pairmint/utils"

	tmprivval "github.com/tendermint/tendermint/privval"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	// Flags
	tmkms bool

	// The start command starts the pairmint application.
	startCmd = &cobra.Command{
		Use:   "start",
		Short: "Start the pairmint application",
		Long:  "",
		Run: func(cmd *cobra.Command, args []string) {
			// Initialize configuration directory.
			configDir := config.GetDir()

			// Create a logger.
			logger := log.New(os.Stderr, "", 0)

			// Configure the PrivValidator.
			pv := privval.NewPairmintFilePV()
			logger.Println("[INFO] pairmint: Loading configuration...")

			if err := pv.Config.Load(); err != nil {
				logger.Printf("[ERR] pairmint: error while loading configuration: %v\n", err)
				os.Exit(1)
			}
			logger.Println("[INFO] pairmint: Successfully loaded configuration. ✓")

			pv.FilePV = tmprivval.LoadOrGenFilePV(pv.Config.FilePV.KeyFilePath, pv.Config.FilePV.StateFilePath)

			// Configure minimum log level for logger.
			logger.SetOutput(&logutils.LevelFilter{
				Levels:   []logutils.LogLevel{"DEBUG", "INFO", "WARN", "ERR"},
				MinLevel: logutils.LogLevel(pv.Config.Init.LogLevel),
				Writer:   os.Stderr,
			})

			// Load the keypair from the pm-identity.key file. The private key is necessary for
			// establishing a secret connection to Tendermint.
			priv, _, err := utils.LoadKeypair(configDir + "/pm-identity.key")
			if err != nil {
				logger.Printf("[ERR] pairmint: error while loading keypair: %v\n", err)
				os.Exit(1)
			}
			logger.Println("[DEBUG] pairmint: Successfully loaded keypair. ✓")

			// Establish a secret connection to the Tendermint validator.
			rwc := connection.NewReadWriteConn()
			logger.Println("[INFO] pairmint: Dialing Tendermint validator...")

			rwc.SecretConn, err = connection.RetrySecretDial("tcp", pv.Config.Init.ValidatorAddr, priv)
			if err != nil {
				logger.Printf("[ERR] pairmint: error while establishing secret connection: %v\n", err)
				os.Exit(1)
			}
			defer rwc.SecretConn.Close()
			logger.Println("[DEBUG] pairmint: Successfully dialed Tendermint validator. ✓")

			rwc.Reader = protoio.NewDelimitedReader(rwc.SecretConn, 64<<10)
			rwc.Writer = protoio.NewDelimitedWriter(rwc.SecretConn)

			// Run the routine for reading and writing messages.
			go pv.Run(rwc, logger)

			// Block until SIGINT is fired.
			osCh := make(chan os.Signal, 1)
			signal.Notify(osCh, os.Interrupt)

			<-osCh
			logger.Println("\nExiting pairmint")
			os.Exit(1)
		},
	}
)

func init() {
	// Add start to the root command.
	rootCmd.AddCommand(startCmd)

	// Add tmkms flag.
	startCmd.Flags().BoolVar(&tmkms, "tmkms", false, "Use the TMKMS as an external PrivValidator process for Pairmint")
	viper.BindPFlag("tmkms", startCmd.Flags().Lookup("tmkms"))
}
