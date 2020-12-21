package privval

import (
	"log"

	"github.com/BlockscapeNetwork/pairmint/config"
	"github.com/BlockscapeNetwork/pairmint/types"
)

var _ types.Pairminter = new(PairmintExternalPV)

// PairmintExternalPV is an external PrivValidator.
type PairmintExternalPV struct {
	// Logger is the logger used to print log messages.
	Logger *log.Logger

	// Config is the node's configuration from the pairmint.toml file.
	Config *config.Config

	// Rank is the node's current rank in the set.
	// All nodes are part of a ranking system: rank #1 is always the
	// signer while the ranks below that serve as backups that gradually
	// move up the ranks if the current signer misses too many blocks in
	// a row.
	Rank int

	// Missed is the counter used to count missed blocks in a row.
	MissedInARow int
}

// Missed adds an entry for a missed block to the frame's list.
// Returns true if the threshold was reached and a rank update
// needs to be done, and false if the threshold has not yet been
// reached.
func (p *PairmintExternalPV) Missed() bool {
	p.MissedInARow++
	if p.MissedInARow == p.Config.Init.Threshold {
		p.MissedInARow = 0
		return true
	}

	return false
}

// Reset resets the frame's missed block counter.
func (p *PairmintExternalPV) Reset() {
	p.MissedInARow = 0
}

// Promote moves the node up one rank. Since there is no explicit
// node demotion, the current rank #1 is automatically demoted to
// the last rank.
func (p *PairmintExternalPV) Promote() {
	if p.Rank > 1 {
		p.Logger.Printf("[INFO] pairmint: Updating rank (%v -> %v)", p.Rank, p.Rank-1)
		p.Rank--
	} else {
		p.Logger.Printf("[INFO] pairmint: Updating rank (%v -> %v)", p.Rank, p.Config.Init.SetSize)
		p.Rank = p.Config.Init.SetSize
	}
}
