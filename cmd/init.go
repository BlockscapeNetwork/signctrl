package cmd

import (
	"fmt"
	"os"

	init_util "github.com/BlockscapeNetwork/signctrl/cmd/init"
	"github.com/BlockscapeNetwork/signctrl/config"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	newPrivval bool
	initCmd    = &cobra.Command{
		Use:   "init",
		Short: "Initializes the SignCTRL node",
		Long:  "Creates the .signctrl/ directory, including a config.toml and a conn.key file",
		Run: func(cmd *cobra.Command, args []string) {
			// Get the config directory.
			cfgDir := config.Dir()

			// Create the config directory.
			if err := os.MkdirAll(cfgDir, config.PermConfigDir); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			fmt.Printf("Created configuration directory (%v)\n", cfgDir)

			// Create the config file.
			if err := init_util.CreateConfigFile(cfgDir); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			// Create the connection key.
			if err := init_util.CreateConnKeyFile(cfgDir); err != nil {
				fmt.Println(err)
				os.Exit(1)
			}

			// Create new priv_validator_key.json and priv_validator_state.json files if --new-pv flag is set.
			if newPrivval {
				if err := init_util.CreateKeyAndStateFiles(cfgDir); err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
			}
		},
	}
)

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().BoolVar(&newPrivval, "new-pv", false, "Creates a new priv_validator_key.json and a priv_validator_state.json file in the configuration directory")
	viper.BindPFlag("new-pv", initCmd.Flags().Lookup("new-pv"))
}
