package privval

import (
	"errors"
	"io"
	"log"

	"github.com/tendermint/tendermint/privval"

	privvalproto "github.com/tendermint/tendermint/proto/tendermint/privval"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/BlockscapeNetwork/pairmint/config"
	"github.com/BlockscapeNetwork/pairmint/connection"
	"github.com/tendermint/tendermint/crypto"
)

var _ Pairminter = new(PairmintFilePV)
var _ tmtypes.PrivValidator = new(PairmintFilePV)

// PairmintFilePV is a wrapper for Tendermint's FilePV.
// Implements the Pairminter and PrivValidator interfaces.
type PairmintFilePV struct {
	// Config is the node's configuration from the pairmint.toml file.
	Config *config.Config

	// MissedInARow is the counter used to count missed blocks in a row.
	MissedInARow int

	// FilePV is Tendermint's file-based signer.
	FilePV *privval.FilePV
}

// NewPairmintFilePV returns a new instance of PairmintFilePV.
func NewPairmintFilePV() *PairmintFilePV {
	return &PairmintFilePV{
		Config:       new(config.Config),
		MissedInARow: 0,
		FilePV:       new(privval.FilePV),
	}
}

// Missed implements the Pairminter interface.
func (p *PairmintFilePV) Missed() error {
	p.MissedInARow++
	if p.MissedInARow == p.Config.Init.Threshold {
		p.MissedInARow = 0
		return errors.New("[ERR] pairmint: too many missed blocks in a row")
	}

	return nil
}

// Reset implements the Pairminter interface.
func (p *PairmintFilePV) Reset() {
	p.MissedInARow = 0
}

// Update implements the Pairminter interface.
func (p *PairmintFilePV) Update() {
	if p.Config.Init.Rank > 1 {
		p.Config.Init.Rank--
	} else {
		p.Config.Init.Rank = p.Config.Init.SetSize
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

// Run runs the routine for the file-based signer.
func (p *PairmintFilePV) Run(rwc *connection.ReadWriteConn, logger *log.Logger) {
	for {
		msg := privvalproto.Message{}
		if _, err := rwc.Reader.ReadMsg(&msg); err != nil {
			if err == io.EOF {
				// Prevent the console log from being spammed with EOF errors.
				continue
			}
			logger.Printf("[ERR] pairmint: error while reading message: %v\n", err)
		}

		if err := p.HandleMessage(&msg, rwc); err != nil {
			logger.Printf("[ERR] pairmint: %v\n", err)
		}
	}
}
