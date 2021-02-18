package config

import (
	"errors"
	"fmt"
	"net"
	"os"
	"regexp"
	"strings"

	"github.com/spf13/viper"
)

// InitConfig defines the initialization parameters for a SignCTRL node.
type InitConfig struct {
	// LogLevel defines the log levels for SignCTRL logs: DEBUG, INFO, WARN, ERR.
	LogLevel string `mapstructure:"log_level"`

	// SetSize determines the fixed size of the SignCTRL node set.
	// The current signer needs to know the set size in order to know which
	// rank to fall back to if it fails.
	SetSize int `mapstructure:"set_size"`

	// Threshold determines the threshold value of consecutive missed block
	// signatures for rank updates.
	Threshold int `mapstructure:"threshold"`

	// Rank determines the SignCTRL node's initial rank on startup.
	Rank int `mapstructure:"rank"`

	// ValidatorListenAddr is the TCP socket address the Tendermint validator
	// listens on for an external PrivValidator process. SignCTRL dials this
	// address to establish a connection to the validator and receive signing
	// requests.
	ValidatorListenAddr string `mapstructure:"validator_laddr"`

	// ValidatorListenAddrRPC is the TCP socket address the validator's RPC
	// server listens on.
	ValidatorListenAddrRPC string `mapstructure:"validator_laddr_rpc"`
}

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

// Config defines the structure of the config.toml file.
type Config struct {
	// Init defines the section for the initialization parameters.
	Init InitConfig `mapstructure:"init"`

	// FilePV defines the section for the file-based signer's file paths.
	FilePV FilePVConfig `mapstructure:"file_pv"`
}

// InitDir creates the SignCTRL configuration directory according
// to the SIGNCTRL_CONFIG_DIR environment variable.
func InitDir(configDir string) error {
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		if err := os.MkdirAll(configDir, 0744); err != nil {
			return err
		}
		fmt.Printf("Created .signctrl/ configuration directory at %v\n", strings.TrimSuffix(configDir, "/.signctrl"))
	}

	return nil
}

// GetDir returns the configuration directory for SignCTRL from the
// SIGNCTRL_CONFIG_DIR environment variable. If the env var is not set
// to a custom directory, it will default to $HOME/.signctrl.
func GetDir() string {
	if os.Getenv("SIGNCTRL_CONFIG_DIR") == "" {
		os.Setenv("SIGNCTRL_CONFIG_DIR", os.Getenv("HOME")+"/.signctrl")
	}

	return os.Getenv("SIGNCTRL_CONFIG_DIR")
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
	if match, _ := regexp.MatchString(`DEBUG|INFO|WARN|ERR`, c.Init.LogLevel); !match {
		errs += "\tlog_level must be either DEBUG, INFO, WARN or ERR\n"
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
		return errors.New(errs)
	}

	return nil
}

// validateFilePVConfig validates the FilePVConfig.
func (c *Config) validateFilePVConfig() error {
	errs := ""
	if c.FilePV.ChainID == "" {
		errs += "\tchain_id must not be empty\n"
	}
	if c.FilePV.KeyFilePath == "" {
		c.FilePV.KeyFilePath = GetDir() + "/priv_validator_key.json"
	}
	if c.FilePV.StateFilePath == "" {
		c.FilePV.StateFilePath = GetDir() + "/priv_validator_state.json"
	}
	if _, err := os.Stat(c.FilePV.KeyFilePath); err != nil {
		errs += "\tkey_file_path does not exist\n"
	}
	if _, err := os.Stat(c.FilePV.StateFilePath); err != nil {
		errs += "\tstate_file_path does not exist\n"
	}
	if errs != "" {
		return errors.New(errs)
	}

	return nil
}

// validate validates the entire configuration.
func (c *Config) validate() error {
	errs := ""
	if err := c.validateInitConfig(); err != nil {
		errs += err.Error()
	}
	if err := c.validateFilePVConfig(); err != nil {
		errs += err.Error()
	}
	if errs != "" {
		return errors.New("\n" + errs)
	}

	return nil
}

// Load loads and validates the configuration parameters for the SignCTRL node.
func (c *Config) Load() error {
	viper.SetConfigName("config")
	viper.AddConfigPath(os.Getenv("SIGNCTRL_CONFIG_DIR"))

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
