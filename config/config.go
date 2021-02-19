package config

import (
	"errors"
	"fmt"
	"net"
	"os"
	"regexp"
	"strings"

	"github.com/hashicorp/logutils"
	"github.com/spf13/viper"
)

const (
	// File is the full file name of the configuration file.
	File = "config.toml"
)

var (
	// LogLevels defines the loglevels for SignCTRL logs.
	LogLevels = []logutils.LogLevel{"DEBUG", "INFO", "WARN", "ERR"}
)

// Init defines the configuration parameters that SignCTRL needs on initialization.
type Init struct {
	// LogLevel determines the minimum log level for SignCTRL logs.
	// Can be DEBUG, INFO, WARN or ERR.
	LogLevel string `mapstructure:"log_level"`

	// SetSize determines the number of validators in the SignCTRL set.
	SetSize int `mapstructure:"set_size"`

	// Threshold determines the threshold value of missed blocks in a row that triggers
	// a rank update in the SignCTRL set.
	Threshold int `mapstructure:"threshold"`

	// Rank determines the validator's rank and therefore whether it has permission
	// to sign votes/proposals or not.
	Rank int `mapstructure:"rank"`

	// ValidatorListenAddress is the TCP socket address the validator listens on for
	// an external PrivValidator process. SignCTRL dials this address to establish a
	// connection with the validator.
	ValidatorListenAddress string `mapstructure:"validator_laddr"`

	// ValidatorListenAddressRPC is the TCP socket address the validator's RPC server
	// listens on.
	ValidatorListenAddressRPC string `mapstructure:"validator_laddr_rpc"`
}

// PrivValidator defines the types of private validators that sign incoming sign
// requests.
type PrivValidator struct {
	// ChainID is the chain that the validator validates for.
	ChainID string `mapstructure:"chain_id"`
}

// Config defines the structure of SignCTRL's configuration file.
type Config struct {
	// Init defines the [init] section of the configuration file.
	Init Init `mapstructure:"init"`

	// Privval defines the [privval] section of the configuration file.
	Privval PrivValidator `mapstructure:"privval"`
}

// Dir returns the configuration directory in use. It is always set in the following
// order:
//
// 1) Custom environment variable $SIGNCTRL_CONFIG_DIR
// 2) $HOME/.signctrl
// 3) Current working directory
//
// If one is not set, the directory falls back to the next one.
func Dir() string {
	if dir := os.Getenv("SIGNCTRL_CONFIG_DIR"); dir != "" {
		return dir
	} else if dir, err := os.UserHomeDir(); err == nil {
		return dir + "/.signctrl"
	}

	return "."
}

// FilePath returns the absolute path to the configuration file.
func FilePath(cfgDir string) string {
	return cfgDir + "/" + File
}

// logLevelsToRegExp returns a regular expression for the validation of log levels.
func logLevelsToRegExp(levels *[]logutils.LogLevel) string {
	var regExp string
	for i, lvl := range *levels {
		regExp += string(lvl)
		if i < len(*levels)-1 {
			regExp += "|"
		}
	}

	return regExp
}

// validateInit validates the configuration's init section.
func validateInit(c *Config) error {
	var errs string
	if match, _ := regexp.MatchString(logLevelsToRegExp(&LogLevels), c.Init.LogLevel); !match {
		errs += fmt.Sprintf("\tlog_level must be one of the following: %v\n", LogLevels)
	}
	if c.Init.SetSize < 2 {
		errs += "\tset_size must be 2 or higher\n"
	}
	if c.Init.Threshold < 1 {
		errs += "\tthreshold must be 1 or higher\n"
	}
	if c.Init.Rank < 1 {
		errs += "\trank must be 1 or higher\n"
	}
	if !strings.HasPrefix(c.Init.ValidatorListenAddress, "tcp://") {
		errs += "\tvalidator_laddr is missing the protocol\n"
	} else {
		host, _, err := net.SplitHostPort(strings.Trim(c.Init.ValidatorListenAddress, "tcp://"))
		if err != nil {
			errs += "\tvalidator_laddr is not in the host:port format\n"
		} else {
			if ip := net.ParseIP(host); ip == nil {
				errs += "\tvalidator_laddr is not a valid IPv4 address\n"
			}
		}
	}
	if !strings.HasPrefix(c.Init.ValidatorListenAddressRPC, "tcp://") {
		errs += "\tvalidator_laddr_rpc is missing the protocol\n"
	} else {
		host, _, err := net.SplitHostPort(strings.Trim(c.Init.ValidatorListenAddressRPC, "tcp://"))
		if err != nil {
			errs += "\tvalidator_laddr_rpc is not in the host:port format\n"
		} else {
			if ip := net.ParseIP(host); ip == nil {
				errs += "\tvalidator_laddr_rpc is not a valid IPv4 address\n"
			}
		}
	}
	if errs != "" {
		return errors.New(errs)
	}

	return nil
}

// validatePrivValidator validates the configuration's privval section.
func validatePrivValidator(c *Config) error {
	var errs string
	if c.Privval.ChainID == "" {
		errs += "\tchain_id must not be empty\n"
	}
	if errs != "" {
		return errors.New(errs)
	}

	return nil
}

// validate validates the configuration.
func validate(c *Config) error {
	var errs string
	if err := validateInit(c); err != nil {
		errs += err.Error()
	}
	if err := validatePrivValidator(c); err != nil {
		errs += err.Error()
	}
	if errs != "" {
		return errors.New(errs)
	}

	return nil
}

// Load loads and validates the configuration file.
func Load() (c *Config, err error) {
	if err = viper.ReadInConfig(); err != nil {
		return nil, err
	}
	if err = viper.Unmarshal(&c); err != nil {
		return nil, err
	}
	if err = validate(c); err != nil {
		return nil, err
	}

	return c, nil
}
