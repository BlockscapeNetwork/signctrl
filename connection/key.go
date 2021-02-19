package connection

import (
	"encoding/base64"
	"io/ioutil"
	"os"

	tm_ed25519 "github.com/tendermint/tendermint/crypto/ed25519"
)

const (
	// KeyFile is the full file name of the connection key.
	KeyFile = "conn.key"

	// PermConnKeyFile determines the default file permisssions for the connection
	// key file.
	PermConnKeyFile = os.FileMode(0700)
)

// KeyFilePath returns the absolute path to the connection key file.
func KeyFilePath(cfgDir string) string {
	return cfgDir + "/" + KeyFile
}

// LoadConnKey loads the connection key from the connection key file.
func LoadConnKey(cfgDir string) (tm_ed25519.PrivKey, error) {
	encSeed, err := ioutil.ReadFile(KeyFilePath(cfgDir))
	if err != nil {
		return nil, err
	}

	decSeed := make([]byte, tm_ed25519.PrivateKeySize)
	if _, err := base64.StdEncoding.Decode(decSeed, encSeed); err != nil {
		return nil, err
	}

	return decSeed, nil
}

// CreateBase64ConnKey creates a base64-encoded connection key.
func CreateBase64ConnKey(cfgDir string) error {
	connKey := tm_ed25519.GenPrivKey()
	encKey := make([]byte, base64.StdEncoding.EncodedLen(tm_ed25519.PrivateKeySize))
	base64.StdEncoding.Encode(encKey, connKey)

	return ioutil.WriteFile(KeyFilePath(cfgDir), encKey, PermConnKeyFile)
}
