package init

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/BlockscapeNetwork/signctrl/config"
	"github.com/BlockscapeNetwork/signctrl/connection"
	"github.com/BlockscapeNetwork/signctrl/privval"
	tm_privval "github.com/tendermint/tendermint/privval"
)

// confirm asks the user for confirmation on file creation, like when a file is about
// to be overwritten. It handles "y" and "yes" for approval, and "", "n" and "no" for
// denial.
func confirm() bool {
	for {
		input, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			fmt.Printf("parsing error: %v", err)
			continue
		}

		switch strings.TrimSuffix(strings.ToLower(input), "\n") {
		case "y", "yes":
			return true
		case "", "n", "no":
			return false
		default:
			fmt.Printf("Unknown answer. Please try again. [y/N]: ")
			continue
		}
	}
}

// CreateConfigFile creates the configuration file in the specified configuration
// directory. In case it already exists, the user is asked to decide whether it
// should be overwritten or not.
func CreateConfigFile(cfgDir string) error {
	if _, err := os.Stat(config.FilePath(cfgDir)); !os.IsNotExist(err) {
		fmt.Printf("Found existing %v at %v. Do you want to overwrite it? [y/N]: ", config.File, cfgDir)
		if confirm() {
			os.Remove(config.FilePath(cfgDir))
			if err := config.Create(cfgDir); err != nil {
				return err
			}
			fmt.Printf("Created %v at %v\n", config.File, cfgDir)
		}
	} else {
		if err := config.Create(cfgDir); err != nil {
			return err
		}
		fmt.Printf("Created %v at %v\n", config.File, cfgDir)
	}

	return nil
}

// CreateConnKeyFile creates the connection key file in the specified configuration
// directory. In case it already exists, the user is asked to decide whether it should
// be overwritten or not.
func CreateConnKeyFile(cfgDir string) error {
	if _, err := os.Stat(connection.KeyFilePath(cfgDir)); !os.IsNotExist(err) {
		fmt.Printf("Found existing %v at %v. Do you want to overwrite it? [y/N]: ", connection.KeyFile, cfgDir)
		if confirm() {
			os.Remove(connection.KeyFilePath(cfgDir))
			if err := connection.CreateBase64ConnKey(cfgDir); err != nil {
				return err
			}
			fmt.Printf("Created new %v at %v\n", connection.KeyFile, cfgDir)
		}
	} else {
		if err := connection.CreateBase64ConnKey(cfgDir); err != nil {
			return err
		}
		fmt.Printf("Created %v at %v\n", connection.KeyFile, cfgDir)
	}

	return nil
}

// CreateKeyAndStateFiles creates the priv_validator_key.json and priv_validator_state.json
// in the specified configuration directory. In case it already exists, the user is asked
// to decide whether it should be overwritten or not.
func CreateKeyAndStateFiles(cfgDir string) error {
	if _, err := os.Stat(privval.KeyFilePath(cfgDir)); !os.IsNotExist(err) {
		fmt.Printf("Found existing priv_validator_key.json at %v. Do you want to overwrite it? [y/N]: ", cfgDir)
		if confirm() {
			os.Remove(privval.KeyFilePath(cfgDir))
			os.Remove(privval.StateFilePath(cfgDir))
			tm_privval.LoadOrGenFilePV(privval.KeyFilePath(cfgDir), privval.StateFilePath(cfgDir))
			fmt.Printf("Created new priv_validator_key.json and priv_validator_state.json at %v\n", cfgDir)
		}
	} else {
		tm_privval.LoadOrGenFilePV(privval.KeyFilePath(cfgDir), privval.StateFilePath(cfgDir))
		fmt.Printf("Created priv_validator_key.json and priv_validator_state.json at %v\n", cfgDir)
	}

	return nil
}
