package privval

// SignCtrled defines the functionality of a validator that monitors
// the blockchain for missed blocks in a row and keeps their rank up
// to date.
type SignCtrled interface {
	// Missed increments the internal counter for missed blocks in a
	// row. Once the threshold of too many missed blocks in a row is
	// exceeded, it throws an error.
	Missed() error

	// Reset resets the counter for missed blocks in a row.
	Reset()

	// Update performs a rank update, moving all backup nodes up one
	// rank and the current signer to the last rank.
	Update()
}
