package privval

import "errors"

var (
	// ErrMissingPubKey is thrown if there is no priv_validator_key.json
	// in the configuration directory and thus no pubkey is to be found.
	ErrMissingPubKey = errors.New("no pubkey found")

	// ErrNoSigner is thrown if the validator is currently ranked #2 or lower
	// and is therefore denied signing permissions.
	ErrNoSigner = errors.New("validator has no permission to sign votes/proposals")

	// ErrUninitialized is thrown if the node has not yet been initialized in
	// terms of missing a config.toml and the conn.key.
	ErrUninitialized = errors.New("SignCTRL is not initialized")

	// ErrTooManyMissedBlocks is thrown if the node exceeds the threshold of
	// too many missed blocks in a row.
	ErrTooManyMissedBlocks = errors.New("too many missed blocks in a row")

	// ErrCatchingUp is thrown if the validator is catching up with the
	// global blockchain state.
	ErrCatchingUp = errors.New("validator is catching up")

	// ErrNoCommitSigs is thrown if the validators /commit endpoint is
	// not available and no commitsigs can be retrieved.
	ErrNoCommitSigs = errors.New("couldn't get commitsigs from validator")

	// ErrWrongChainID is thrown if an incoming request is for a different
	// chainid than the one specified in the config.toml file.
	ErrWrongChainID = errors.New("wrong chainid")
)
