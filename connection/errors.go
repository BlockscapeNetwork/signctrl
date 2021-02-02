package connection

import "errors"

var (
	// ErrNoCommitSigs is thrown if the result in the JSON response
	// for the commit RPC is nil. This can be due to passing in an
	// invalid/non-existent height into the query.
	ErrNoCommitSigs = errors.New("no commit signatures found for height ")
)
