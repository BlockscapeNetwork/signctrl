package utils

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
)

// loadAndDecodeBase64Seed loads the base64-encoded seed located at the given file path
// and returns the decoded seed bytes.
func loadAndDecodeBase64Seed(filepath string) ([]byte, error) {
	// Load base64-encoded seed from file.
	encSeed, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	// Decode seed.
	decSeed := make([]byte, ed25519.SeedSize)
	if _, err := base64.StdEncoding.Decode(decSeed, encSeed); err != nil {
		return nil, err
	}

	return decSeed, nil
}

// LoadKeypair generates an ed25519 keypair from the seed located at the given file path.
func LoadKeypair(filepath string) (ed25519.PrivateKey, ed25519.PublicKey, error) {
	// Load decoded seed bytes.
	seed, err := loadAndDecodeBase64Seed(filepath)
	if err != nil {
		return nil, nil, err
	}

	// Prepare and return private and public key.
	privKey := ed25519.PrivateKey(ed25519.NewKeyFromSeed(seed))
	pubKey := privKey.Public().(ed25519.PublicKey)

	return privKey, pubKey, nil
}

// writeBase64Seed base64-encodes the given seed bytes and writes them to the pm-identity.key
// file in the `SIGNCTRL_CONFIG_DIR` directory with restricted file permissions (0700).
func writeBase64Seed(filepath string, seed []byte) error {
	encSeed := make([]byte, base64.StdEncoding.EncodedLen(len(seed)))
	base64.StdEncoding.Encode(encSeed, seed)
	if err := ioutil.WriteFile(filepath, encSeed, 0700); err != nil {
		return err
	}

	return nil
}

// GenSeed generates an ed25519 seed of 32 bytes which is the identity of the
// SignCTRL node. This seed corresponds to RFC 8032's private key.
func GenSeed(filepath string) error {
	// GenerateKey creates an RFC 8032 ed25519 private key of 64 bytes which is
	// used for signing messages. Under the hood, the private key is a SHA-512
	// hash of the seed.
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return err
	}
	if err := writeBase64Seed(filepath, priv.Seed()); err != nil {
		return err
	}

	fmt.Printf("Created new pm-identity.key seed at %v\n", os.Getenv("SIGNCTRL_CONFIG_DIR"))

	return nil
}
