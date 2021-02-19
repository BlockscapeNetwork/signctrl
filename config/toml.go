package config

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

const (
	// PermConfigDir determines the default file permissions for the configuration
	// directory.
	PermConfigDir = os.FileMode(0744)

	// PermConfigToml determines the default file permissions for the configuration
	// file.
	PermConfigToml = os.FileMode(0644)
)

// Section is a custom type for specific sections in the configuration file.
type Section uint8

const (
	// InitSection defines the [init] section of the configuration file.
	InitSection Section = iota

	// PrivvalSection defines the [privval] section of the configuration file.
	PrivvalSection
)

// goPath returns the $GOPATH directory. It is retrieved using the 'go env GOPATH'
// command.
func goPath() string {
	gopath, err := exec.Command("go", "env", "GOPATH").Output()
	if err != nil {
		return ""
	}

	return strings.TrimSuffix(string(gopath), "\n")
}

// templateDir gets the directory for the template of a configuration section.
func templateDir(temp Section) string {
	path := goPath() + "/src/github.com/BlockscapeNetwork/signctrl/config/templates/"
	switch temp {
	case InitSection:
		return path + "init.toml"
	case PrivvalSection:
		return path + "privval.toml"
	}

	return ""
}

// Create writes configuration templates to the configuration file at the specified
// configuration directory. The init and privval sections are created by default.
func Create(cfgDir string, sections ...Section) error {
	var cfg bytes.Buffer

	initBytes, err := ioutil.ReadFile(templateDir(InitSection))
	if err != nil {
		return fmt.Errorf("%v\nRun: 'git pull https://github.com/BlockscapeNetwork/signctrl.git' to get the configuration template", err)
	}
	if _, err := cfg.Write(initBytes); err != nil {
		return err
	}

	privvalBytes, err := ioutil.ReadFile(templateDir(PrivvalSection))
	if err != nil {
		return fmt.Errorf("%v\nRun: 'git pull https://github.com/BlockscapeNetwork/signctrl.git' to get the configuration template", err)
	}
	if _, err := cfg.Write(privvalBytes); err != nil {
		return err
	}

	if err := ioutil.WriteFile(FilePath(cfgDir), cfg.Bytes(), PermConfigToml); err != nil {
		return err
	}

	return err
}
