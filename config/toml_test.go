package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreate(t *testing.T) {
	err := Create(".")
	defer os.Remove("./config.toml")
	assert.NoError(t, err)
}
