package config

import (
	"fmt"
	"io/ioutil"
	"os"

	tm_json "github.com/tendermint/tendermint/libs/json"
)

const (
	// StateFile is the full file name of the file that persists the validator's
	// last state.
	StateFile = "signctrl_state.json"

	// PermStateFile determines the default file permissions for the
	// signctrl_state.json file.
	PermStateFile = os.FileMode(0644)
)

// State defines the contents of the signctrl_state.json file.
type State struct {
	LastSignedHeight int64 `json:"last_signed_height"`
	LastRank         int   `json:"last_rank"`
}

// StateFilePath returns the absolute path to the signctrl_state.json file.
func StateFilePath(cfgDir string) string {
	return cfgDir + "/" + StateFile
}

// validateState validates the contents of the signctrl_state.json file.
func validateState(s State) error {
	var errs string
	if s.LastSignedHeight < 1 {
		errs += "\tlast_signed_height in signctrl_state.json must be 1 or higher\n"
	}
	if s.LastRank < 1 {
		errs += "\tlast_rank in signctrl_state.json must be 1 or higher\n"
	}
	if errs != "" {
		return fmt.Errorf(errs)
	}

	return nil
}

// LoadState loads the contents of the signctrl_state.json file and returns them.
func LoadState(cfgDir string) (State, error) {
	if _, err := os.Stat(StateFilePath(cfgDir)); os.IsNotExist(err) {
		return State{}, err
	}

	bytes, err := ioutil.ReadFile(StateFilePath(cfgDir))
	if err != nil {
		return State{}, err
	}

	var s State
	if err := tm_json.Unmarshal(bytes, &s); err != nil {
		return State{}, err
	}
	if err := validateState(s); err != nil {
		return State{}, err
	}

	return s, nil
}

// SaveState saves the current state to the signctrl_state.json file.
func SaveState(cfgDir string, s State) error {
	lrFile, err := tm_json.MarshalIndent(&State{
		LastRank:         s.LastRank,
		LastSignedHeight: s.LastSignedHeight,
	}, "", "\t")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(StateFilePath(cfgDir), lrFile, PermStateFile)
}
