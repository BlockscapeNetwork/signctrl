package cmd

import (
	"fmt"
	"os"

	builder "github.com/BlockscapeNetwork/pairmint/cmd/init"
	"github.com/BlockscapeNetwork/pairmint/config"
	"github.com/BlockscapeNetwork/pairmint/utils"
	"github.com/spf13/cobra"
)

// The init command creates a pairmint.toml configuration file and a
// secret seed used for establishing a secret connection to Tendermint
// and an external PrivValidator process.
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize pairmint",
	Long:  "Create a pairmint.toml configuration template and generate a pm-identity.key seed.",
	Run: func(cmd *cobra.Command, args []string) {
		// If no custom configuration directory is set, always use the default one.
		if os.Getenv("PAIRMINT_CONFIG_DIR") == "" {
			os.Setenv("PAIRMINT_CONFIG_DIR", os.Getenv("HOME")+"/.pairmint")
		}

		// Initialize the config directory.
		configDir := os.Getenv("PAIRMINT_CONFIG_DIR")
		if err := config.InitConfigDir(configDir); err != nil {
			fmt.Printf("couldn't initialize configuration directory: %v\n", err.Error())
			os.Exit(1)
		}

		// Build the configuration template.
		if err := builder.BuildConfigTemplate(configDir); err != nil {
			fmt.Printf("couldn't build config template: %v\n", err.Error())
			os.Exit(1)
		}

		// Generate the pm-identity.key seed if there isn't already one.
		seedPath := os.Getenv("PAIRMINT_CONFIG_DIR") + "/pm-identity.key"
		if _, err := os.Stat(seedPath); os.IsNotExist(err) {
			if err := utils.GenSeed(seedPath); err != nil {
				fmt.Printf("couldn't generate seed: %v\n", err.Error())
				os.Exit(1)
			}
		} else {
			fmt.Printf("Found existing pm-identity.key seed at %v\n", configDir)
		}
	},
}

func init() {
	// Add init to the root command.
	rootCmd.AddCommand(initCmd)
}
