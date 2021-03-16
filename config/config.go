package config

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

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

// Base defines the base configuration parameters for SignCTRL.
type Base struct {
	// LogLevel determines the minimum log level for SignCTRL logs.
	// Can be DEBUG, INFO, WARN or ERR.
	LogLevel string `mapstructure:"log_level"`

	// SetSize determines the number of validators in the SignCTRL set.
	SetSize int `mapstructure:"set_size"`

	// Threshold determines the threshold value of missed blocks in a row that
	// triggers a rank update in the SignCTRL set.
	Threshold int `mapstructure:"threshold"`

	// StartRank determines the validator's rank on startup and therefore whether it
	// has permission to sign votes/proposals or not.
	StartRank int `mapstructure:"start_rank"`

	// ValidatorListenAddress is the TCP socket address the validator listens on for
	// an external PrivValidator process. SignCTRL dials this address to establish a
	// connection with the validator.
	ValidatorListenAddress string `mapstructure:"validator_laddr"`

	// ValidatorListenAddressRPC is the TCP socket address the validator's RPC server
	// listens on.
	ValidatorListenAddressRPC string `mapstructure:"validator_laddr_rpc"`

	// RetryDialAfter is the time after which SignCTRL assumes it lost connection to
	// the validator and retries dialing it.
	RetryDialAfter string `mapstructure:"retry_dial_after"`
}

// validate validates the configuration's base section.
func (b Base) validate() error {
	var errs string
	if match, _ := regexp.MatchString(logLevelsToRegExp(LogLevels), b.LogLevel); !match {
		errs += fmt.Sprintf("\tlog_level must be one of the following: %v\n", LogLevels)
	}
	if b.SetSize < 2 {
		errs += "\tset_size must be 2 or higher\n"
	}
	if b.Threshold < 1 {
		errs += "\tthreshold must be 1 or higher\n"
	}
	if b.StartRank < 1 {
		errs += "\tstart_rank must be 1 or higher\n"
	}
	protocol := regexp.MustCompile(`(tcp|unix)://`).FindString(b.ValidatorListenAddress)
	if protocol == "" {
		errs += "\tvalidator_laddr is missing the protocol\n"
	} else if protocol == "tcp://" {
		host, _, err := net.SplitHostPort(strings.TrimPrefix(b.ValidatorListenAddress, protocol))
		if err != nil {
			errs += "\tvalidator_laddr is not in the host:port format\n"
		} else {
			if ip := net.ParseIP(host); ip == nil {
				errs += "\tvalidator_laddr is not a valid IPv4 address\n"
			}
		}
	} else if protocol == "unix://" {
		if !strings.HasSuffix(b.ValidatorListenAddress, ".sock") {
			errs += "\nvalidator_laddr is not a unix domain socket address\n"
		}
	}
	if !strings.HasPrefix(b.ValidatorListenAddressRPC, "tcp://") {
		errs += "\tvalidator_laddr_rpc is missing the protocol\n"
	} else {
		host, _, err := net.SplitHostPort(strings.Trim(b.ValidatorListenAddressRPC, "tcp://"))
		if err != nil {
			errs += "\tvalidator_laddr_rpc is not in the host:port format\n"
		} else {
			if ip := net.ParseIP(host); ip == nil {
				errs += "\tvalidator_laddr_rpc is not a valid IPv4 address\n"
			}
		}
	}
	if b.RetryDialAfter == "" {
		errs += "\tretry_dial_after must not be empty\n"
	} else {
		time := regexp.MustCompile(`[1-9][0-9]+`).FindString(b.RetryDialAfter)
		if time == "" {
			errs += "\tretry_dial_after is missing the time\n"
		}
		timeUnit := regexp.MustCompile(`s|m|h`).FindString(b.RetryDialAfter)
		if timeUnit == "" {
			errs += "\tretry_dial_after is missing the unit of time\n"
		}
	}
	if errs != "" {
		return errors.New(errs)
	}

	return nil
}

// PrivValidator defines the types of private validators that sign incoming sign
// requests.
type PrivValidator struct {
	// ChainID is the chain that the validator validates for.
	ChainID string `mapstructure:"chain_id"`
}

// validate validates the configuration's privval section.
func (p PrivValidator) validate() error {
	var errs string
	if p.ChainID == "" {
		errs += "\tchain_id must not be empty\n"
	}
	if errs != "" {
		return errors.New(errs)
	}

	return nil
}

// Config defines the structure of SignCTRL's configuration file.
type Config struct {
	// Base defines the [base] section of the configuration file.
	Base Base `mapstructure:"base"`

	// Privval defines the [privval] section of the configuration file.
	Privval PrivValidator `mapstructure:"privval"`
}

// validate validates the configuration.
func (c Config) validate() error {
	var errs string
	if err := c.Base.validate(); err != nil {
		errs += err.Error()
	}
	if err := c.Privval.validate(); err != nil {
		errs += err.Error()
	}
	if errs != "" {
		return errors.New(errs)
	}

	return nil
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
	return filepath.Join(cfgDir, File)
}

// GetRetryDialTime converts the string representation of RetryDialAfter into
// time.Duration and returns it.
func GetRetryDialTime(timeString string) time.Duration {
	t := regexp.MustCompile(`0|[1-9][0-9]*`).FindString(timeString)
	tConv, _ := strconv.Atoi(t)

	tUnit := regexp.MustCompile(`s|m|h`).FindString(timeString)
	switch tUnit {
	case "s":
		return time.Duration(tConv) * time.Second
	case "m":
		return time.Duration(tConv) * time.Minute
	case "h":
		return time.Duration(tConv) * time.Hour
	}

	return 0
}

// logLevelsToRegExp returns a regular expression for the validation of log levels.
func logLevelsToRegExp(levels *[]logutils.LogLevel) string {
	regExp := ""
	maxLevels := len(*levels) - 1
	for i, lvl := range *levels {
		regExp += string(lvl)
		if i < maxLevels {
			regExp += "|"
		}
	}

	return regExp
}

// Load loads and validates the configuration file.
func Load() (c Config, err error) {
	if err = viper.ReadInConfig(); err != nil {
		return Config{}, err
	}
	if err = viper.Unmarshal(&c); err != nil {
		return Config{}, err
	}
	if err = c.validate(); err != nil {
		return Config{}, err
	}

	return c, nil
}
