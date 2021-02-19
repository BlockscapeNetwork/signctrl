package cmd

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/BlockscapeNetwork/signctrl/config"
	"github.com/BlockscapeNetwork/signctrl/connection"
	"github.com/BlockscapeNetwork/signctrl/privval"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	tm_privval "github.com/tendermint/tendermint/privval"
)

var (
	startCmd = &cobra.Command{
		Use:   "start",
		Short: "Starts the SignCTRL node",
		Run: func(cmd *cobra.Command, args []string) {
			// Get the config directory.
			cfgDir := config.Dir()

			// Load the config into memory.
			cfg, err := config.Load()
			if err != nil {
				fmt.Printf("Couldn't load config.toml:\n%v", err)
				os.Exit(1)
			}

			// Initialize a new SCFilePV.
			pv := privval.NewSCFilePV(
				log.New(os.Stderr, "", 0),
				cfg,
				tm_privval.LoadOrGenFilePV(privval.KeyFilePath(cfgDir), privval.StateFilePath(cfgDir)),
			)

			// Load the connection key from the config directory.
			connKey, err := connection.LoadConnKey(cfgDir)
			if err != nil {
				fmt.Printf("Couldn't load conn.key: %v", err)
				os.Exit(1)
			}

			// Dial the validator.
			secretConn, err := connection.RetrySecretDialTCP(
				cfg.Init.ValidatorListenAddress,
				connKey,
				pv.Logger,
			)
			if err != nil {
				fmt.Printf("Couldn't dial validator: %v", err)
				os.Exit(1)
			}
			defer secretConn.Close()

			// Start main goroutine for message handling.
			// Wait for SIGINT or SIGTERM to terminate SignCTRL.
			<-pv.Start(secretConn)
		},
	}
)

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.AddCommand(startCmd)
}

func initConfig() {
	cfgParts := strings.Split(config.File, ".")

	viper.SetConfigName(cfgParts[0])
	viper.SetConfigType(cfgParts[1])

	viper.AddConfigPath("$SIGNCTRL_CONFIG_DIR")
	viper.AddConfigPath("$HOME/.signctrl")
	viper.AddConfigPath(".")
}
