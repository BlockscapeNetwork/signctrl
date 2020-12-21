package privval

import (
	"log"
	"os"

	"github.com/tendermint/tendermint/privval"

	"github.com/hashicorp/logutils"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/BlockscapeNetwork/pairmint/types"

	"github.com/BlockscapeNetwork/pairmint/config"
	"github.com/tendermint/tendermint/crypto"
)

var _ types.Pairminter = new(PairmintFilePV)
var _ tmtypes.PrivValidator = new(PairmintFilePV)

// PairmintFilePV is a wrapper for Tendermint's FilePV.
type PairmintFilePV struct {
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

	// MissedInARow is the counter used to count missed blocks in a row.
	MissedInARow int

	// FilePV is Tendermint's file-based PrivValidator.
	FilePV *privval.FilePV
}

// NewPairmintFilePV returns a new instance of PairmintFilePV.
func NewPairmintFilePV() (*PairmintFilePV, error) {
	pm := &PairmintFilePV{
		Logger:       log.New(os.Stderr, "", 0),
		Config:       new(config.Config),
		Rank:         0,
		MissedInARow: 0,
		FilePV:       new(privval.FilePV),
	}

	if err := pm.Config.Load(); err != nil {
		return pm, err
	}

	pm.Logger.SetOutput(&logutils.LevelFilter{
		Levels:   []logutils.LogLevel{"DEBUG", "INFO", "WARN", "ERR"},
		MinLevel: logutils.LogLevel(pm.Config.Init.LogLevel),
		Writer:   os.Stderr,
	})

	pm.Rank = pm.Config.Init.Rank
	pm.FilePV = privval.LoadOrGenFilePV(pm.Config.FilePV.KeyFilePath, pm.Config.FilePV.StateFilePath)

	return pm, nil
}

// Missed adds an entry for a missed block to the frame's list.
// Returns true if the threshold was reached and a rank update
// needs to be done, and false if the threshold has not yet been
// reached.
func (p *PairmintFilePV) Missed() bool {
	p.MissedInARow++
	if p.MissedInARow == p.Config.Init.Threshold {
		p.MissedInARow = 0
		return true
	}

	return false
}

// Reset resets the frame's missed block counter.
func (p *PairmintFilePV) Reset() {
	p.MissedInARow = 0
}

// Promote moves the node up one rank. Since there is no explicit
// node demotion, the current rank #1 is automatically demoted to
// the last rank.
func (p *PairmintFilePV) Promote() {
	if p.Rank > 1 {
		p.Logger.Printf("[INFO] pairmint: Updating rank (%v -> %v)", p.Rank, p.Rank-1)
		p.Rank--
	} else {
		p.Logger.Printf("[INFO] pairmint: Updating rank (%v -> %v)", p.Rank, p.Config.Init.SetSize)
		p.Rank = p.Config.Init.SetSize
	}
}

// GetPubKey returns the public key of the validator.
// Implements the PrivValidator interface.
func (p *PairmintFilePV) GetPubKey() (crypto.PubKey, error) {
	return p.FilePV.GetPubKey()
}

// SignVote signs a canonical representation of the vote, along with the
// chainID. Implements the PrivValidator interface.
func (p *PairmintFilePV) SignVote(chainID string, vote *tmproto.Vote) error {
	return p.FilePV.SignVote(chainID, vote)
}

// SignProposal signs a canonical representation of the proposal, along with
// the chainID. Implements the PrivValidator interface.
func (p *PairmintFilePV) SignProposal(chainID string, proposal *tmproto.Proposal) error {
	return p.FilePV.SignProposal(chainID, proposal)
}
