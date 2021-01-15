package privval

import "errors"

var (
	// ErrMissingPubKey is thrown if there is no priv_validator_key.json
	// in the configuration directory and thus no pubkey is to be found.
	ErrMissingPubKey = errors.New("no pubkey found")

	// ErrNoSigner is thrown if the validator is currently ranked #2 or lower
	// and is therefore denied signing permissions.
	ErrNoSigner = errors.New("validator has no permission to sign votes/proposals")

	// ErrUninitialized is thrown if pairmint has not yet been initialized in
	// terms of missing a pairmint.toml and the pm-identity.key.
	ErrUninitialized = errors.New("pairmint is not initialized")
)
