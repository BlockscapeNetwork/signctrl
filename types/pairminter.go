package types

import (
	"log"
	"os"

	"github.com/BlockscapeNetwork/pairmint/config"
	"github.com/hashicorp/logutils"
)

// Pairminter defines the attributes of a pairminter node.
type Pairminter struct {
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

// NewPairminter creates a new, fully configured Pairminter instance.
func NewPairminter() (*Pairminter, error) {
	pm := &Pairminter{
		Logger:       log.New(os.Stderr, "", 0),
		Config:       &config.Config{},
		Rank:         0,
		MissedInARow: 0,
	}

	if err := pm.Config.Load(); err != nil {
		return nil, err
	}
	pm.Rank = pm.Config.Init.Rank

	pm.Logger.SetOutput(&logutils.LevelFilter{
		Levels:   []logutils.LogLevel{"DEBUG", "INFO", "WARN", "ERR"},
		MinLevel: logutils.LogLevel(pm.Config.Init.LogLevel),
		Writer:   os.Stderr,
	})

	return pm, nil
}

// Missed adds an entry for a missed block to the frame's list.
// Returns true if the threshold was reached and a rank update
// needs to be done, and false if the threshold has not yet been
// reached.
func (p *Pairminter) Missed() bool {
	p.MissedInARow++
	if p.MissedInARow == p.Config.Init.Threshold {
		p.MissedInARow = 0
		return true
	}

	return false
}

// Reset resets the frame's missed block counter.
func (p *Pairminter) Reset() {
	p.MissedInARow = 0
}

// Promote moves the node up one rank. Since there is no explicit
// node demotion, the current rank #1 is automatically demoted to
// the last rank.
func (p *Pairminter) Promote() {
	if p.Rank > 1 {
		p.Logger.Printf("[INFO] pairmint: Updating rank (%v -> %v)", p.Rank, p.Rank-1)
		p.Rank--
	} else {
		p.Logger.Printf("[INFO] pairmint: Updating rank (%v -> %v)", p.Rank, p.Config.Init.SetSize)
		p.Rank = p.Config.Init.SetSize
	}
}
