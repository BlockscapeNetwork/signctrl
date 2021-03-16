package connection

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKeyFilePath(t *testing.T) {
	path := KeyFilePath("/tmp")
	assert.Equal(t, "/tmp/conn.key", path)
}

func TestCreateAndLoadConnKey(t *testing.T) {
	// Fail to load conn.key.
	key, err := LoadConnKey(".")
	assert.Nil(t, key)
	assert.Error(t, err)

	// Succeed loading conn.key.
	err = CreateBase64ConnKey(".")
	assert.NoError(t, err)
	defer os.Remove("./conn.key")

	key, err = LoadConnKey(".")
	assert.NotNil(t, key)
	assert.NoError(t, err)
}
