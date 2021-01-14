package privval

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/tendermint/tendermint/libs/protoio"
	"github.com/tendermint/tendermint/privval"

	p2pconn "github.com/tendermint/tendermint/p2p/conn"
	cryptoproto "github.com/tendermint/tendermint/proto/tendermint/crypto"
	privvalproto "github.com/tendermint/tendermint/proto/tendermint/privval"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/BlockscapeNetwork/pairmint/config"
	"github.com/BlockscapeNetwork/pairmint/utils"
	"github.com/tendermint/tendermint/crypto"
)

var _ Pairminter = new(PairmintFilePV)
var _ tmtypes.PrivValidator = new(PairmintFilePV)

// PairmintFilePV is a wrapper for Tendermint's FilePV.
// Implements the Pairminter and PrivValidator interfaces.
type PairmintFilePV struct {
	// SecretConn holds the secret connection for communication between
	// Pairmint and Tendermint.
	SecretConn *p2pconn.SecretConnection

	// Reader is used to read from the TCP stream.
	Reader protoio.ReadCloser

	// Writer is used to write to the TCP stream.
	Writer protoio.WriteCloser

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
		SecretConn:   new(p2pconn.SecretConnection),
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

// --------------------------------------------------------------------------

// Run runs the routine for the file-based signer.
func (p *PairmintFilePV) Run(secretconn *p2pconn.SecretConnection, logger *log.Logger) {
	for {
		msg := privvalproto.Message{}
		if _, err := p.Reader.ReadMsg(&msg); err != nil {
			if err == io.EOF {
				// Prevent the console log from being spammed with EOF errors.
				continue
			}
			logger.Printf("[ERR] pairmint: error while reading message: %v\n", err)
		}

		switch msg.GetSum().(type) {
		case *privvalproto.Message_PingRequest:
			logger.Println("[DEBUG] pairmint: Received PingRequest")

			if _, err := p.Writer.WriteMsg(utils.WrapMsg(&privvalproto.PingResponse{})); err != nil {
				logger.Printf("[ERR] pairmint: error while writing PingResponse: %v\n", err)
				continue
			}

		case *privvalproto.Message_PubKeyRequest:
			req := msg.GetPubKeyRequest()
			logger.Printf("[DEBUG] pairmint: Received PubKeyRequest: %v\n", req)

			// Get the pubkey from the priv_validator_key.json file.
			pubkey, err := p.GetPubKey()
			if err != nil {
				logger.Printf("[ERR] pairmint: couldn't get pubkey: %v\n", err)
				os.Exit(1)
			}

			if _, err = p.Writer.WriteMsg(utils.WrapMsg(&privvalproto.PubKeyResponse{
				PubKey: cryptoproto.PublicKey{
					Sum: &cryptoproto.PublicKey_Ed25519{
						Ed25519: pubkey.Bytes(),
					},
				}})); err != nil {
				logger.Printf("[ERR] pairmint: error while writing PubKeyResponse: %v\n", err)
				continue
			}

		case *privvalproto.Message_SignVoteRequest:
			req := msg.GetSignVoteRequest()
			logger.Printf("[DEBUG] pairmint: Received SignVoteRequest: %v\n", req)

			// TODO: Sign vote if node is primary, and reply with signed vote.
			// TODO: Else, reply with an error.

			// Sign vote.
			if err := p.FilePV.SignVote(p.Config.FilePV.ChainID, req.Vote); err != nil {
				logger.Printf("[ERR] pairmint: couldn't sign vote for height %v: %v\n", req.Vote.Height, err)
				continue
			}

			if _, err := p.Writer.WriteMsg(utils.WrapMsg(&privvalproto.SignedVoteResponse{Vote: *req.Vote})); err != nil {
				logger.Printf("[ERR] pairmint: error while writing SignedVoteResponse: %v\n", err)
				continue
			}

		case *privvalproto.Message_SignProposalRequest:
			req := msg.GetSignProposalRequest()
			logger.Printf("[DEBUG] pairmint: Received SignProposalRequest: %v\n", req)

			// TODO: Sign proposal if node is primary, and reply with signed vote.
			// TODO: Else, reply with an error.

			// Sign proposal.
			if err := p.FilePV.SignProposal(p.Config.FilePV.ChainID, req.Proposal); err != nil {
				logger.Printf("[ERR] pairmint: couldn't sign proposal for height %v: %v\n", req.Proposal.Height, err)
				continue
			}

			if _, err := p.Writer.WriteMsg(utils.WrapMsg(&privvalproto.SignedProposalResponse{Proposal: *req.Proposal})); err != nil {
				logger.Printf("[ERR] pairmint: error while writing SignedProposalResponse: %v\n", err)
				continue
			}

		default:
			panic(fmt.Sprintf("unknown sum type: %T", msg.GetSum()))
		}
	}
}
