package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	builder "github.com/BlockscapeNetwork/pairmint/cmd/init"
	"github.com/BlockscapeNetwork/pairmint/config"
	"github.com/BlockscapeNetwork/pairmint/utils"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	tmprivval "github.com/tendermint/tendermint/privval"
)

var (
	// Flags
	keypair bool

	// The init command creates a pairmint.toml configuration file and a
	// secret seed used for establishing a secret connection to Tendermint
	// and an external PrivValidator process.
	initCmd = &cobra.Command{
		Use:   "init",
		Short: "Initialize pairmint",
		Long:  "Create a pairmint.toml configuration template and generate a pm-identity.key seed.",
		Run: func(cmd *cobra.Command, args []string) {
			// Get the configuration directory.
			configDir := config.GetDir()

			// Initialize the config directory.
			if err := config.InitDir(configDir); err != nil {
				fmt.Printf("couldn't initialize configuration directory: %v\n", err)
				os.Exit(int(syscall.SIGHUP))
			}

			// Build the configuration template.
			if err := builder.BuildConfigTemplate(configDir); err != nil {
				fmt.Printf("couldn't build config template: %v\n", err)
				os.Exit(int(syscall.SIGHUP))
			}

			// Generate the pm-identity.key seed if there isn't already one.
			seedPath := configDir + "/pm-identity.key"
			if _, err := os.Stat(seedPath); os.IsNotExist(err) {
				if err := utils.GenSeed(seedPath); err != nil {
					fmt.Printf("couldn't generate seed: %v\n", err)
					os.Exit(int(syscall.SIGHUP))
				}
			} else {
				fmt.Printf("Found existing pm-identity.key seed at %v\n", configDir)
			}

			// Generate new key and state files if --keypair flag is set.
			if keypair {
				keyPath := configDir + "/priv_validator_key.json"
				statePath := configDir + "/priv_validator_state.json"

				if _, err := os.Stat(keyPath); !os.IsNotExist(err) {
					fmt.Printf("Found existing priv_validator_key.json at %v. Are you sure you want to overwrite it?\n[y|N]: ", configDir)
					reader := bufio.NewReader(os.Stdin)
					input, err := reader.ReadString('\n')
					if err != nil {
						fmt.Printf("error while reading input: %v\n", err)
						os.Exit(int(syscall.SIGHUP))
					}

					if input == "\n" || strings.ToLower(input) == "n\n" {
						fmt.Println("Creation of key and state file canceled.")
						os.Exit(int(syscall.SIGHUP))
					} else if strings.ToLower(input) == "y\n" {
						os.Remove(keyPath)
						os.Remove(statePath)
						tmprivval.LoadOrGenFilePV(keyPath, statePath)
						fmt.Printf("Created new priv_validator_key.json and priv_validator_state.json at %v\n", configDir)
					}
				} else {
					tmprivval.LoadOrGenFilePV(keyPath, statePath)
					fmt.Printf("Created new priv_validator_key.json and priv_validator_state.json at %v\n", configDir)
				}
			}
		},
	}
)

func init() {
	// Add init to the root command.
	rootCmd.AddCommand(initCmd)

	// Add flags.
	initCmd.Flags().BoolVar(&keypair, "keypair", false, "Generate a new priv_validator_key.json and priv_validator_state.json in the $PAIRMINT_CONFIG_DIR directory")
	viper.BindPFlag("keypair", initCmd.Flags().Lookup("keypair"))
}
