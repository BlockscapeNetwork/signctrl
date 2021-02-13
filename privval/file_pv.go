package privval

import (
	"io"
	"log"
	"os"

	"github.com/tendermint/tendermint/privval"

	privvalproto "github.com/tendermint/tendermint/proto/tendermint/privval"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/BlockscapeNetwork/signctrl/config"
	"github.com/BlockscapeNetwork/signctrl/connection"
	tmcrypto "github.com/tendermint/tendermint/crypto"
)

var _ SignCtrled = new(SCFilePV)
var _ tmtypes.PrivValidator = new(SCFilePV)

// SCFilePV is a wrapper for Tendermint's FilePV.
// Implements the SignCtrled and PrivValidator interfaces.
type SCFilePV struct {
	// Logger is the logger used to log SignCTRL messages.
	Logger *log.Logger

	// Config is the node's configuration from the config.toml file.
	Config *config.Config

	// MissedInARow is the counter used to count missed blocks in a row.
	MissedInARow int

	// CounterUnlocked is the toggle used for allowing the node to count
	// missed blocks in a row.
	CounterUnlocked bool

	// FilePV is Tendermint's file-based signer.
	FilePV *privval.FilePV

	// CurrentHeight keeps track of the current height based on the
	// messages SignCTRL receives from the validator. It is used to keep
	// track of which height the commitsigs were retrieved at.
	CurrentHeight int64
}

// NewSCFilePV returns a new instance of SCFilePV.
func NewSCFilePV() *SCFilePV {
	return &SCFilePV{
		Logger:          new(log.Logger),
		Config:          new(config.Config),
		MissedInARow:    0,
		CounterUnlocked: false,
		FilePV:          new(privval.FilePV),
		CurrentHeight:   1, // Start at genesis block height
	}
}

// Missed implements the SignCtrled interface.
func (p *SCFilePV) Missed() error {
	if !p.CounterUnlocked {
		p.Logger.Printf("[INFO] signctrl: Haven't found commitsig from rank 1 since having synced up")
		return nil
	}

	p.MissedInARow++
	p.Logger.Printf("[INFO] signctrl: Missed a block (%v/%v)\n", p.MissedInARow, p.Config.Init.Threshold)

	if p.MissedInARow == p.Config.Init.Threshold {
		p.Reset()
		return ErrTooManyMissedBlocks
	}

	return nil
}

// Reset implements the SignCtrled interface.
func (p *SCFilePV) Reset() {
	if p.MissedInARow != 0 {
		p.MissedInARow = 0
		p.Logger.Printf("[INFO] signctrl: Reset counter for missed blocks in a row (%v/%v)\n", p.MissedInARow, p.Config.Init.Threshold)
	}
}

// Update implements the SignCtrled interface.
func (p *SCFilePV) Update() {
	// TODO: Uncomment this if statement when signer rank demotion gets implemented
	// if p.Config.Init.Rank == 1 {
	// 	p.Config.Init.Rank = p.Config.Init.SetSize
	// 	p.Logger.Printf("[INFO] signctrl: Demoted validator (rank #1 -> #%v)\n", p.Config.Init.Rank)
	// }

	p.Config.Init.Rank--
	p.Reset()
	p.Logger.Printf("[INFO] signctrl: Promoted validator (rank #%v -> #%v)\n", p.Config.Init.Rank+1, p.Config.Init.Rank)
}

// GetPubKey returns the public key of the validator.
// Implements the PrivValidator interface.
func (p *SCFilePV) GetPubKey() (tmcrypto.PubKey, error) {
	return p.FilePV.GetPubKey()
}

// SignVote signs a canonical representation of the vote, along with the
// chainID. Implements the PrivValidator interface.
func (p *SCFilePV) SignVote(chainID string, vote *tmproto.Vote) error {
	return p.FilePV.SignVote(chainID, vote)
}

// SignProposal signs a canonical representation of the proposal, along with
// the chainID. Implements the PrivValidator interface.
func (p *SCFilePV) SignProposal(chainID string, proposal *tmproto.Proposal) error {
	return p.FilePV.SignProposal(chainID, proposal)
}

// Run runs the routine for the file-based signer.
func (p *SCFilePV) Run(rwc *connection.ReadWriteConn, pubkey tmcrypto.PubKey, sigCh chan os.Signal) {
	p.Logger.Println("[INFO] signctrl: Running SignCTRL daemon...")

	for {
		select {
		case <-sigCh:
			p.Logger.Println("[INFO] signctrl: Terminating SignCTRL...")
			return

		default:
			msg := privvalproto.Message{}
			if _, err := rwc.Reader.ReadMsg(&msg); err != nil {
				if err == io.EOF {
					// Prevent the console log from being spammed with EOF errors.
					continue
				}
				p.Logger.Printf("[ERR] signctrl: error while reading message: %v\n", err)
			}

			if err := p.HandleMessage(&msg, pubkey, rwc); err != nil {
				p.Logger.Printf("[ERR] signctrl: couldn't handle message: %v\n", err)
			}
		}
	}
}
