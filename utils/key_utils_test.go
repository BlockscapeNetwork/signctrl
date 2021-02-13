package utils

import (
	"io/ioutil"
	"testing"

	"os"
)

func TestLoadKeypair(t *testing.T) {
	ioutil.WriteFile("./pm-identity.key", []byte("i0E6lqsN1aw9KQjgLG+c7YpoJ0oPTYrttwk0aExZkZE="), 0644)
	defer os.Remove("./pm-identity.key")

	priv, pub, err := LoadKeypair("./pm-identity.key")
	if err != nil {
		t.Errorf("Expected err to be nil, instead got: %v", err)
	}
	if len(priv) != 64 {
		t.Errorf("Expected private key to be 64 bytes long, instead got: %v", len(priv))
	}
	if len(pub) != 32 {
		t.Errorf("Expected public key to be 32 bytes long, instead got: %v", len(pub))
	}
}

func TestGenSeed(t *testing.T) {
	defer os.Remove("./pm-identity.key")
	if err := GenSeed("./pm-identity.key"); err != nil {
		t.Errorf("Expected pm-identity.key file to be created, instead got err: %v", err)
	}
}
