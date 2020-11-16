package pairmint

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

// Config defines the configuration parameters for a pairmint node.
type Config struct {
	// The amount of pairminters in the queue.
	// This number also tells the current signer which rank to
	// fall back to if it fails.
	QueueSize uint `json:"queue_size"`

	// The pairminters rank on startup/initialization.
	InitRank uint `json:"init_rank"`

	// The size of the block frame that is monitored for signed
	// and missed blocks.
	FrameSize uint `json:"frame_size"`
}

// load loads the pairmint configuration from the directory specified in
// the environment variable PAIRMINT_CONFIG_FILE. By default, this directory
// is $HOME/.pairmint.
func (c *Config) load() error {
	if os.Getenv("PAIRMINT_CONFIG_FILE") == "" {
		pwd, _ := os.Getwd()
		os.Setenv("PAIRMINT_CONFIG_FILE", fmt.Sprintf("%v/.pairmint", pwd))
	}

	configDir := fmt.Sprintf("%v/pairmint.json", os.Getenv("PAIRMINT_CONFIG_FILE"))
	config, err := os.Open(configDir)
	if err != nil {
		return err
	}
	defer config.Close()

	bytes, err := ioutil.ReadAll(config)
	if err != nil {
		return err
	}
	if err = json.Unmarshal(bytes, &c); err != nil {
		return err
	}
	return nil
}
