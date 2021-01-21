package config

import (
	"fmt"
	"net"
	"os"
	"regexp"
	"strings"

	"github.com/spf13/viper"
)

// InitConfig defines the initialization parameters for a pairmint node.
type InitConfig struct {
	// LogLevel defines the log levels for pairmint logs: DEBUG, INFO, WARN, ERR.
	LogLevel string `mapstructure:"log_level"`

	// SetSize determines the fixed size of the pairminter set.
	// The current signer needs to know the set size in order to know which
	// rank to fall back to if it fails.
	SetSize int `mapstructure:"set_size"`

	// Threshold determines the threshold value of consecutive missed block
	// signatures for rank updates.
	Threshold int `mapstructure:"threshold"`

	// Rank determines the pairminters initial rank on startup.
	Rank int `mapstructure:"rank"`

	// ValidatorListenAddr is the TCP socket address the Tendermint validator
	// listens on for an external PrivValidator process. Pairmint dials this
	// address to establish a connection to the validator and receive signing
	// requests.
	ValidatorListenAddr string `mapstructure:"validator_laddr"`

	// ValidatorListenAddrRPC is the TCP socket address the validator's RPC
	// server listens on.
	ValidatorListenAddrRPC string `mapstructure:"validator_laddr_rpc"`
}

// // ExtPVConfig defines address of an external PrivValidator process for Pairmint
// // to connect to.
// type ExtPVConfig struct {
// 	// PrivValidatorListenAddr is the TCP socket address to listen on for
// 	// connections from an external PrivValidator process.
// 	PrivValidatorListenAddr string `mapstructure:"priv_validator_laddr"`
// }

// FilePVConfig defines file paths for the file-based signer.
type FilePVConfig struct {
	// The chain ID the FilePV signs votes/proposals for.
	ChainID string `mapstructure:"chain_id"`

	// KeyFilePath is the absolute path to the priv_validator_key.json file
	// needed to run the file-based signer.
	KeyFilePath string `mapstructure:"key_file_path"`

	// StateFilePath is the absolute path to the priv_validator_state.json
	// file needed to run the file-based signer.
	StateFilePath string `mapstructure:"state_file_path"`
}

// Config defines the structure of the pairmint.toml file.
type Config struct {
	// Init defines the section for the initialization parameters.
	Init InitConfig `mapstructure:"init"`

	// // Tmkms defines the section for tmkms configuration parameters.
	// ExtPV ExtPVConfig `mapstructure:"ext_pv"`

	// FilePV defines the section for the file-based signer's file paths.
	FilePV FilePVConfig `mapstructure:"file_pv"`
}

// InitDir creates the pairmint configuration directory according
// to the `PAIRMINT_CONFIG_DIR` encironment variable.
func InitDir(configDir string) error {
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		if err := os.MkdirAll(configDir, 0744); err != nil {
			return err
		}
		fmt.Printf("Created .pairmint/ configuration directory at %v\n", strings.TrimSuffix(configDir, "/.pairmint"))
	}

	return nil
}

// GetDir returns the configuration directory for pairmint from the
// PAIRMINT_CONFIG_DIR environment variable. If the env var is not set
// to a custom directory, it will default to $HOME/.pairmint.
func GetDir() string {
	if os.Getenv("PAIRMINT_CONFIG_DIR") == "" {
		os.Setenv("PAIRMINT_CONFIG_DIR", os.Getenv("HOME")+"/.pairmint")
	}

	return os.Getenv("PAIRMINT_CONFIG_DIR")
}

// validateInitConfig validates the InitConfig.
func (c *Config) validateInitConfig() error {
	errs := ""
	if c.Init.SetSize < 2 {
		errs += "\tset_size must be 2 or higher\n"
	}
	if c.Init.Threshold <= 0 {
		errs += "\tthreshold must be greater than 0\n"
	}
	if c.Init.Rank <= 0 {
		errs += "\trank must be greater than 0"
		if c.Init.Rank > c.Init.SetSize {
			errs += " and smaller than set_size"
		}
		errs += "\n"
	}
	if c.Init.LogLevel == "" {
		if match, _ := regexp.MatchString(`DEBUG|INFO|WARN|ERR`, c.Init.LogLevel); !match {
			errs += "\tlog_level must be either DEBUG, INFO, WARN or ERR\n"
		}
	}
	if c.Init.ValidatorListenAddr != "" {
		host, _, err := net.SplitHostPort(c.Init.ValidatorListenAddr)
		if err != nil {
			errs += "\tvalidator_laddr is not in the host:port format\n"
		} else {
			if ip := net.ParseIP(host); ip == nil {
				errs += "\tvalidator_laddr is not a valid IPv4\n"
			}
		}
	}
	if c.Init.ValidatorListenAddrRPC != "" {
		host, _, err := net.SplitHostPort(c.Init.ValidatorListenAddrRPC)
		if err != nil {
			errs += "\tvalidator_laddr_rpc is not in the host:port format\n"
		} else {
			if ip := net.ParseIP(host); ip == nil {
				errs += "\tvalidator_laddr_rpc is not a valid IPv4\n"
			}
		}
	}

	if errs != "" {
		return fmt.Errorf("%v", errs)
	}

	return nil
}

// // validateExtPVConfig validates the ExtPVConfig.
// func (c *Config) validateExtPVConfig() error {
// 	errs := ""
// 	if c.ExtPV.PrivValidatorListenAddr != "" {
// 		host, _, err := net.SplitHostPort(c.ExtPV.PrivValidatorListenAddr)
// 		if err != nil {
// 			errs += "\tpriv_validator_laddr is not in the host:port format\n"
// 		} else {
// 			if ip := net.ParseIP(host); ip == nil {
// 				errs += "\tpriv_validator_laddr is not a valid IPv4\n"
// 			}
// 		}
// 	}

// 	if errs != "" {
// 		return fmt.Errorf("%v", errs)
// 	}

// 	return nil
// }

// validateFilePVConfig validates the FilePVConfig.
func (c *Config) validateFilePVConfig() error {
	errs := ""
	if c.FilePV.ChainID == "" {
		errs += "\tchain_id must not be empty\n"
	}
	if keyFile, err := os.Stat(c.FilePV.KeyFilePath); err != nil && !keyFile.IsDir() {
		errs += "\tkey_file_path is not a valid path\n"
	}
	if stateFile, err := os.Stat(c.FilePV.StateFilePath); err != nil && !stateFile.IsDir() {
		errs += "\tstate_file_path is not a valid path\n"
	}

	if errs != "" {
		return fmt.Errorf("%v", errs)
	}

	return nil
}

// validate validates the entire configuration.
func (c *Config) validate() error {
	errs := ""
	if err := c.validateInitConfig(); err != nil {
		errs += err.Error()
	}
	// if err := c.validateExtPVConfig(); err != nil {
	// 	errs += err.Error()
	// }
	if err := c.validateFilePVConfig(); err != nil {
		errs += err.Error()
	}

	if errs != "" {
		return fmt.Errorf("invalid config:\n%v", errs)
	}

	return nil
}

// Load loads and validates the configuration parameters for the pairmint node.
func (c *Config) Load() error {
	viper.SetConfigName("pairmint")
	viper.AddConfigPath(os.Getenv("PAIRMINT_CONFIG_DIR"))

	if err := viper.ReadInConfig(); err != nil {
		return err
	}
	if err := viper.Unmarshal(&c); err != nil {
		return err
	}
	if err := c.validate(); err != nil {
		return err
	}

	return nil
}
