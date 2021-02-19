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
	newKey bool
	// external bool

	initCmd = &cobra.Command{
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
			fmt.Printf("Created %v\n", cfgDir)

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

			// Create new private validator key and state files if --new-key flag is set.
			if newKey {
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
	initCmd.Flags().BoolVar(&newKey, "new-key", false, "Generates a new private validator key along with a state file in the config directory")
	viper.BindPFlag("new-key", initCmd.Flags().Lookup("new-key"))
}
