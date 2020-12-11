package config

import (
	"fmt"
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
	// The current signer needs to know the set size in order to know which rank to fall
	// back to if it fails.
	SetSize int `mapstructure:"set_size"`

	// Threshold determines the threshold value of consecutive missed block signatures for
	// rank updates.
	Threshold int `mapstructure:"threshold"`

	// Rank determines the pairminters initial rank on startup.
	Rank int `mapstructure:"rank"`
}

// ConnectionConfig defines tcp addresses pairmint keeps a connection to.
type ConnectionConfig struct {
	// ValidatorAddr is the TCP socket address of the Tendermint validator node
	// for Pairmint to connect to.
	ValidatorAddr string `mapstructure:"validator_addr"`

	// PrivValidatorListenAddr is the TCP socket address to listen on for connections
	// from an external PrivValidator process.
	PrivValidatorListenAddr string `mapstructure:"priv_validator_laddr"`
}

// Config defines the structure of the pairmint.toml file.
type Config struct {
	// Init defines the section for the initialization parameters.
	Init InitConfig `mapstructure:"init"`

	// Connection defines the section for addresses.
	Connection ConnectionConfig `mapstructure:"connection"`
}

// InitConfigDir creates the pairmint configuration directory according
// to the `PAIRMINT_CONFIG_DIR` encironment variable.
func InitConfigDir(configDir string) error {
	if _, err := os.Stat(configDir); os.IsNotExist(err) {
		if err := os.MkdirAll(configDir, 0744); err != nil {
			return err
		}
		fmt.Printf("Created .pairmint/ configuration directory at %v\n", strings.TrimSuffix(configDir, "/.pairmint"))
	}

	return nil
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
	if errs != "" {
		return fmt.Errorf("%v", errs)
	}

	return nil
}

// validateConnectionConfig validates the ConnectionConfig.
func (c *Config) validateConnectionConfig() error {
	errs := ""
	if strings.HasPrefix(c.Connection.ValidatorAddr, "tcp://") {
		errs += "\tvalidator_addr must start with prefix tcp://\n"
	}
	if strings.HasPrefix(c.Connection.PrivValidatorListenAddr, "tcp://") {
		errs += "\tpriv_validator_laddr must start with prefix tcp://\n"
	}

	// TODO: Properly validate the addresses! Maybe with regex?

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
	if err := c.validateConnectionConfig(); err != nil {
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
