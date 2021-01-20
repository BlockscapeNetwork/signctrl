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

			// Configure the PrivValidator.
			pv := privval.NewPairmintFilePV()
			pv.Logger = log.New(os.Stderr, "", 0)
			if err := pv.Config.Load(); err != nil {
				pv.Logger.Printf("[ERR] pairmint: error while loading configuration: %v\n", err)
				os.Exit(1)
			}
			pv.FilePV = tmprivval.LoadOrGenFilePV(pv.Config.FilePV.KeyFilePath, pv.Config.FilePV.StateFilePath)

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
				os.Exit(1)
			}

			// Establish a secret connection to the Tendermint validator.
			rwc := connection.NewReadWriteConn()
			rwc.SecretConn, err = connection.RetrySecretDial("tcp", pv.Config.Init.ValidatorAddr, priv, pv.Logger)
			if err != nil {
				pv.Logger.Printf("[ERR] pairmint: error while establishing secret connection: %v\n", err)
				os.Exit(1)
			}
			defer rwc.SecretConn.Close()

			rwc.Reader = protoio.NewDelimitedReader(rwc.SecretConn, 64<<10)
			rwc.Writer = protoio.NewDelimitedWriter(rwc.SecretConn)

			pubkey, err := pv.GetPubKey()
			if err != nil {
				pv.Logger.Printf("[ERR] pairmint: couldn't get privval pubkey: %v", err)
				os.Exit(1)
			}

			// Run the routine for reading and writing messages.
			go pv.Run(rwc, pubkey)

			// Block until SIGINT is fired.
			osCh := make(chan os.Signal, 1)
			signal.Notify(osCh, os.Interrupt)

			<-osCh
			pv.Logger.Println("\n[INFO] pairmint: Exiting pairmint")
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
