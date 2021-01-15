package cmd

import (
	"log"
	"os"

	"github.com/hashicorp/logutils"
	"github.com/tendermint/tendermint/libs/protoio"

	"github.com/BlockscapeNetwork/pairmint/config"
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
			configDir := config.GetConfigDir()  // Get the configuration directory.
			logger := log.New(os.Stderr, "", 0) // Create a logger.
			pv := privval.NewPairmintFilePV()   // Initialize new PairmintFilePV instance.

			// Load the configuration parameters.
			if err := pv.Config.Load(); err != nil {
				logger.Printf("[ERR] pairmint: error while loading configuration: %v\n", err)
				os.Exit(1)
			}
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
			logger.Println("[DEBUG] pairmint: Loaded keypair successfully. ✓")

			// Establish a secret connection to the Tendermint validator.
			logger.Println("[INFO] pairmint: Dialing Tendermint validator...")
			pv.SecretConn, err = utils.RetrySecretDial("tcp", pv.Config.Init.ValidatorAddr, priv)
			if err != nil {
				logger.Printf("[ERR] pairmint: error while establishing secret connection: %v\n", err)
				os.Exit(1)
			}
			defer pv.SecretConn.Close()
			logger.Println("[DEBUG] pairmint: Successfully dialed Tendermint validator. ✓")

			pv.Reader = protoio.NewDelimitedReader(pv.SecretConn, 64<<10)
			pv.Writer = protoio.NewDelimitedWriter(pv.SecretConn)

			// Run the routine for reading and writing messages.
			go pv.Run(pv.SecretConn, logger)

			// Keep the application running.
			for {
				select {}
			}
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
