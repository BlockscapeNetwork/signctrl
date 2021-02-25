package config

import (
	"bytes"
	"embed"
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

var (
	// Embed the init.toml into the SignCTRL binary.
	//go:embed templates/init.toml
	initTemplate embed.FS

	// Embed the privval.toml into the SignCTRL binary.
	//go:embed templates/privval.toml
	privvalTemplate embed.FS
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

// Create writes configuration templates to the configuration file at the specified
// configuration directory. The init and privval sections are created by default.
func Create(cfgDir string, sections ...Section) error {
	var cfg bytes.Buffer

	initBytes, err := initTemplate.ReadFile("templates/init.toml")
	if err != nil {
		return err
	}
	if _, err := cfg.Write(initBytes); err != nil {
		return err
	}

	privvalBytes, err := privvalTemplate.ReadFile("templates/privval.toml")
	if err != nil {
		return err
	}
	if _, err := cfg.Write(privvalBytes); err != nil {
		return err
	}

	if err := ioutil.WriteFile(FilePath(cfgDir), cfg.Bytes(), PermConfigToml); err != nil {
		return err
	}

	return err
}
