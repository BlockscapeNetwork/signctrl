package config

import (
	"os"
	"testing"

	"github.com/spf13/afero"
)

func TestInitDir(t *testing.T) {
	configDir := "./.pairminttest"
	fs := afero.NewOsFs()
	defer fs.RemoveAll(configDir)

	if err := InitDir(configDir); err != nil {
		t.Errorf("Expected err to be nil, instead got: %v", err)
	}
	if _, err := fs.Stat(configDir); os.IsNotExist(err) {
		t.Errorf("Expected %v to have been created, instead it doesn't exist", configDir)
	}
}

func TestGetDir(t *testing.T) {
	defaultDir := os.Getenv("HOME") + "/.pairmint"

	os.Setenv("PAIRMINT_CONFIG_DIR", "")
	if GetDir() != defaultDir {
		t.Errorf("Expected PAIRMINT_CONFIG_DIR to be \"%v\", instead got: %v", defaultDir, os.Getenv("PAIRMINT_CONFIG_DIR"))
	}

	os.Setenv("PAIRMINT_CONFIG_DIR", "/some/random/dir")
	if GetDir() != "/some/random/dir" {
		t.Errorf("Expected PAIRMINT_CONFIG_DIR to be \"/some/random/dir\", instead got: %v", os.Getenv("PAIRMINT_CONFIG_DIR"))
	}
}

func TestValidate(t *testing.T) {
	key := `{
"address": "5BCD69E0178E0E6C6F96F541B265CAE3178611AE",
"pub_key": {
  "type": "tendermint/PubKeyEd25519",
  "value": "KwddNyi18Ta7tPs6xwfM79O3waMn1+aJuB6GyGQjYuY="
},
"priv_key": {
  "type": "tendermint/PrivKeyEd25519",
  "value": "XQpf+QIrfT/3v0yLquLhfJ5dUaQfJ+ScLYoPzjpUuTkrB103KLXxNru0+zrHB8zv07fBoyfX5om4HobIZCNi5g=="
  }
}`
	state := `{
  "height": "0",
  "round": 0,
  "step": 0,
}`

	fs := afero.NewOsFs()
	afero.WriteFile(fs, "./priv_validator_key.json", []byte(key), 0644)
	afero.WriteFile(fs, "./priv_validator_state.json", []byte(state), 0644)
	defer fs.Remove("./priv_validator_key.json")
	defer fs.Remove("./priv_validator_state.json")

	config := &Config{
		Init: InitConfig{
			LogLevel:               "INFO",
			SetSize:                2,
			Threshold:              10,
			Rank:                   1,
			ValidatorListenAddr:    "127.0.0.1:4000",
			ValidatorListenAddrRPC: "127.0.0.1:26657",
		},
		FilePV: FilePVConfig{
			ChainID:       "testchain",
			KeyFilePath:   "./priv_validator_key.json",
			StateFilePath: "./priv_validator_state.json",
		},
	}

	// Valid config.
	if err := config.validate(); err != nil {
		t.Errorf("Expected err to be nil, instead got: %v", err)
	}

	// Invalid loglevel.
	config.Init.LogLevel = "INVALID"
	if err := config.validate(); err == nil {
		t.Errorf("Expected err, instead got nil")
	}
	config.Init.LogLevel = "INFO"

	// Invalid setsize.
	config.Init.SetSize = 0
	if err := config.validate(); err == nil {
		t.Errorf("Expected err, instead got nil")
	}
	config.Init.SetSize = 2

	// Invalid threshold.
	config.Init.Threshold = 0
	if err := config.validate(); err == nil {
		t.Errorf("Expected err, instead got nil")
	}
	config.Init.Threshold = 10

	// Invalid rank.
	config.Init.Rank = 0
	if err := config.validate(); err == nil {
		t.Errorf("Expected err, instead got nil")
	}
	config.Init.Rank = 1

	// Invalid validator listen address.
	config.Init.ValidatorListenAddr = "127.0.0.1"
	if err := config.validate(); err == nil {
		t.Errorf("Expected err, instead got nil")
	}
	config.Init.ValidatorListenAddr = "127.0.0.1:4000"

	// Invalid validator rpc listen address.
	config.Init.ValidatorListenAddrRPC = "127.0.0.1"
	if err := config.validate(); err == nil {
		t.Errorf("Expected err, instead got nil")
	}
	config.Init.ValidatorListenAddrRPC = "127.0.0.1:26657"

	// Invalid chainid.
	config.FilePV.ChainID = ""
	if err := config.validate(); err == nil {
		t.Errorf("Expected err, instead got nil")
	}
	config.FilePV.ChainID = "testchain"

	// Non-existent path to keyfile.
	config.FilePV.KeyFilePath = "/this/path/does/not/exist"
	if err := config.validate(); err == nil {
		t.Errorf("Expected err, instead got nil")
	}
	config.FilePV.KeyFilePath = "./priv_validator_key.json"

	// Non-existent path to statefile.
	config.FilePV.StateFilePath = "/this/path/does/not/exist"
	if err := config.validate(); err == nil {
		t.Errorf("Expected err, instead got nil")
	}
	config.FilePV.StateFilePath = "./priv_validator_state.json"
}
