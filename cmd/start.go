package cmd

import (
	"fmt"
	"os"

	"github.com/BlockscapeNetwork/pairmint/types"
	"github.com/BlockscapeNetwork/pairmint/utils"
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
			// If no custom configuration directory is set, always use the default one.
			if os.Getenv("PAIRMINT_CONFIG_DIR") == "" {
				os.Setenv("PAIRMINT_CONFIG_DIR", os.Getenv("HOME")+"/.pairmint")
			}

			// Create new pairminter instance.
			_, err := types.NewPairminter()
			if err != nil {
				fmt.Printf("couldn't initialize pairminter: %v\nPlease run `pairmint init` to initialize the pairmint node.\n", err.Error())
				os.Exit(1)
			}

			seedPath := os.Getenv("PAIRMINT_CONFIG_DIR") + "/pm-identity.key"

			// Generate seed if it doesn't already exist.
			if _, err := os.Stat(seedPath); os.IsNotExist(err) {
				if err := utils.GenSeed(seedPath); err != nil {
					fmt.Println(err.Error())
					os.Exit(1)
				}
			}

			// TODO: MakeSecretConnection to the TMKMS or tendermint/tendermint/privval/file.go.

			// TODO: Start WebSocket server for Tendermint to connect to.

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
