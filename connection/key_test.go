package connection

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKeyFilePath(t *testing.T) {
	path := KeyFilePath("/key_test_filepath")
	assert.Equal(t, "/key_test_filepath/conn.key", path)
}

func TestCreateAndLoadConnKey(t *testing.T) {
	cfgDir := "./key_test_createandload"
	os.MkdirAll(cfgDir, PermConnKeyFile)
	defer os.RemoveAll(cfgDir)

	// Fail to load conn.key.
	key, err := LoadConnKey(cfgDir)
	assert.Nil(t, key)
	assert.Error(t, err)

	// Succeed loading conn.key.
	err = CreateBase64ConnKey(cfgDir)
	assert.NoError(t, err)

	key, err = LoadConnKey(cfgDir)
	assert.NotNil(t, key)
	assert.NoError(t, err)
}
