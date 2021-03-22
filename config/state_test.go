package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func testState(t *testing.T) *State {
	t.Helper()
	return &State{
		LastHeight: 10,
		LastRank:   1,
	}
}

func TestValidateState(t *testing.T) {
	// Valid State.
	state := testState(t)
	err := state.validate()
	assert.NoError(t, err)

	// Invalid State.LastHeight.
	state.LastHeight = 0
	err = state.validate()
	assert.Error(t, err)
	state.LastHeight = testState(t).LastHeight

	// Invalid State.LastRank.
	state.LastRank = 0
	err = state.validate()
	assert.Error(t, err)
	state.LastRank = testState(t).LastRank
}

func TestStateFilePath(t *testing.T) {
	path := StateFilePath("/tmp")
	assert.Equal(t, "/tmp/signctrl_state.json", path)
}

func TestLoadOrGenState(t *testing.T) {
	// Generate.
	state, err := LoadOrGenState(".")
	defer os.Remove("./signctrl_state.json")
	assert.NotNil(t, state)
	assert.NoError(t, err)

	// Load invalid.
	state, err = LoadOrGenState(".")
	assert.Equal(t, state, State{})
	assert.Error(t, err)

	// Load valid.
	state = *testState(t)
	err = state.Save(".")
	assert.NoError(t, err)

	state, err = LoadOrGenState(".")
	assert.Equal(t, state, *testState(t))
	assert.NoError(t, err)
}
