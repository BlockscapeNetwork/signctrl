package types

// Pairminter defines the functionality of a Pairminter that monitors
// the blockchain for missed blocks in a row and keeps their rank up
// to date.
type Pairminter interface {
	Missed() bool
	Reset()
	Promote()
}
