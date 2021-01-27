package privval

import (
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"

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
	// Logger is the logger used to log pairmint messages.
	Logger *log.Logger

	// Config is the node's configuration from the pairmint.toml file.
	Config *config.Config

	// MissedInARow is the counter used to count missed blocks in a row.
	MissedInARow int

	// FilePV is Tendermint's file-based signer.
	FilePV *privval.FilePV

	// CurrentHeight keeps track of the current height based on the
	// messages pairmint receives from the validator.
	CurrentHeight int64
}

// NewPairmintFilePV returns a new instance of PairmintFilePV.
func NewPairmintFilePV() *PairmintFilePV {
	return &PairmintFilePV{
		Logger:        new(log.Logger),
		Config:        new(config.Config),
		MissedInARow:  0,
		FilePV:        new(privval.FilePV),
		CurrentHeight: 0,
	}
}

// Missed implements the Pairminter interface.
func (p *PairmintFilePV) Missed() error {
	p.MissedInARow++
	p.Logger.Printf("[DEBUG] pairmint: Missed a block (total: %v)\n", p.MissedInARow)

	if p.MissedInARow == p.Config.Init.Threshold {
		p.MissedInARow = 0
		return ErrTooManyMissedBlocks
	}

	return nil
}

// Reset implements the Pairminter interface.
func (p *PairmintFilePV) Reset() {
	p.MissedInARow = 0
	p.Logger.Println("[DEBUG] pairmint: Reset counter for missed blocks in a row")
}

// Update implements the Pairminter interface.
func (p *PairmintFilePV) Update() {
	if p.Config.Init.Rank > 1 {
		p.Config.Init.Rank--
		p.Logger.Printf("[DEBUG] pairmint: Promoted validator from rank %v to %v\n", p.Config.Init.Rank+1, p.Config.Init.Rank)
	} else {
		p.Config.Init.Rank = p.Config.Init.SetSize
		p.Logger.Printf("[DEBUG] pairmint: Demoted validator from rank 1 to %v\n", p.Config.Init.Rank)
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
func (p *PairmintFilePV) Run(rwc *connection.ReadWriteConn, pubkey crypto.PubKey) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	defer close(sigCh)

	for {
		select {
		case <-sigCh:
			p.Logger.Println("[INFO] pairmint: Terminating pairmint...")
			return

		default:
			msg := privvalproto.Message{}
			if _, err := rwc.Reader.ReadMsg(&msg); err != nil {
				if err == io.EOF {
					// Prevent the console log from being spammed with EOF errors.
					continue
				}
				p.Logger.Printf("[ERR] pairmint: error while reading message: %v\n", err)
			}

			if err := p.HandleMessage(&msg, pubkey, rwc); err != nil {
				p.Logger.Printf("[ERR] pairmint: couldn't handle message: %v\n", err)
			}
		}
	}
}
