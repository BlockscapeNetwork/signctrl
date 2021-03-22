package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

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
	LastHeight int64 `json:"last_height"`
	LastRank   int   `json:"last_rank"`
}

// validate validates the contents of the signctrl_state.json file.
func (s State) validate() error {
	var errs string
	if s.LastHeight < 1 {
		errs += "\tlast_height in signctrl_state.json must be 1 or higher\n"
	}
	if s.LastRank < 1 {
		errs += "\tlast_rank in signctrl_state.json must be 1 or higher\n"
	}
	if errs != "" {
		return fmt.Errorf(errs)
	}

	return nil
}

// StateFilePath returns the absolute path to the signctrl_state.json file.
func StateFilePath(cfgDir string) string {
	return filepath.Join(cfgDir, StateFile)
}

// LoadOrGenState loads the contents of the signctrl_state.json file and returns them
// if it exists, or generetas a new one.
func LoadOrGenState(cfgDir string) (State, error) {
	if _, err := os.Stat(StateFilePath(cfgDir)); os.IsNotExist(err) {
		state := State{
			LastHeight: 1,
			LastRank:   0,
		}
		if err := state.Save(cfgDir); err != nil {
			return State{}, err
		}

		return state, nil
	}

	bytes, err := ioutil.ReadFile(StateFilePath(cfgDir))
	if err != nil {
		return State{}, err
	}

	var s State
	if err := tm_json.Unmarshal(bytes, &s); err != nil {
		return State{}, err
	}
	if err := s.validate(); err != nil {
		return State{}, err
	}

	return s, nil
}

// Save saves the current state to the signctrl_state.json file.
func (s *State) Save(cfgDir string) error {
	lrFile, err := tm_json.MarshalIndent(&State{
		LastRank:   s.LastRank,
		LastHeight: s.LastHeight,
	}, "", "\t")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(StateFilePath(cfgDir), lrFile, PermStateFile)
}
