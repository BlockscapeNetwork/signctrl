package privval

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/BlockscapeNetwork/pairmint/connection"
	"github.com/tendermint/tendermint/crypto"
	cryptoproto "github.com/tendermint/tendermint/proto/tendermint/crypto"
	privvalproto "github.com/tendermint/tendermint/proto/tendermint/privval"
)

// handlePingRequest handles incoming ping requests.
func (p *PairmintFilePV) handlePingRequest(rwc *connection.ReadWriteConn) error {
	p.Logger.Println("[DEBUG] pairmint: Write PingResponse")
	if _, err := rwc.Writer.WriteMsg(wrapMsg(&privvalproto.PingResponse{})); err != nil {
		return err
	}

	return nil
}

// handlePubKeyRequest handles incoming public key requests.
func (p *PairmintFilePV) handlePubKeyRequest(req *privvalproto.PubKeyRequest, pubkey crypto.PubKey, rwc *connection.ReadWriteConn) error {
	resp := &privvalproto.PubKeyResponse{
		PubKey: cryptoproto.PublicKey{
			Sum: &cryptoproto.PublicKey_Ed25519{
				Ed25519: pubkey.Bytes(),
			},
		},
	}

	p.Logger.Printf("[DEBUG] pairmint: Write PubKeyResponse: %v\n", pubkey.Address())
	if _, err := rwc.Writer.WriteMsg(wrapMsg(resp)); err != nil {
		return err
	}

	return nil
}

// handleSignVoteRequest handles incoming signing requests for prevotes and precommits.
func (p *PairmintFilePV) handleSignVoteRequest(req *privvalproto.SignVoteRequest, pubkey crypto.PubKey, rwc *connection.ReadWriteConn) error {
	// Prepare empty vote response.
	resp := &privvalproto.SignedVoteResponse{}

	// Only check the commitsigs once for each block height.
	// Since p.CurrentHeight is initialized to 1, the check for the genesis block is
	// skipped as there is no previous commit to be fetched.
	if req.Vote.Height > p.CurrentHeight {
		// Retrieve last height's commit from the /commit endpoint of the validator.
		commitsigs, err := connection.GetCommitSigs(req.Vote.Height - 1)
		if err != nil {
			p.Logger.Printf("[ERR] pairmint: couldn't get commitsigs: %v\n", err)
			resp.Error = &privvalproto.RemoteSignerError{Description: err.Error()}

			// Send error to Tendermint that the commitsigs could not be retrieved.
			// In this case, pairmint can't know whether it is safe to sign or not,
			// so it won't sign the message.
			if _, err := rwc.Writer.WriteMsg(wrapMsg(resp)); err != nil {
				return err
			}

			return ErrNoCommitSigs
		}

		p.Logger.Printf("[DEBUG] pairmint: GET /commit?height=%v: %v\n", req.Vote.Height-1, commitsigs)
		p.CurrentHeight = req.Vote.Height

		// Check if the last commit contains our validator's signature.
		if hasSignedCommit(pubkey.Address(), commitsigs) {
			p.Logger.Printf("[DEBUG] pairmint: Found signature from %v in commitsigs from height %v\n", pubkey.Address().String(), req.Vote.Height-1)
			p.Reset()
		} else {
			p.Logger.Printf("[ERR] pairmint: no commitsig from %v for block height %v\n", pubkey.Address().String(), req.Vote.Height-1)

			// None of the commitsigs had an entry with our validator's address and
			// a signature in them which means that this block was missed.
			if err := p.Missed(); err != nil {
				p.Logger.Println("[ERR] pairmint: too many missed blocks in a row, updating ranks...")

				// If an error is thrown it means that the threshold of too many missed
				// blocks in a row has been exceeded. Now, a rank update is done in order
				// to replace the signer.
				p.Update()
				p.Reset()
			}
		}
	}

	// Check if the validator has permission to sign the vote.
	if p.Config.Init.Rank == 1 {
		p.Logger.Printf("[DEBUG] pairmint: Validator is ranked #1, signing %v...\n", req.Vote.Type)

		// Sign the vote.
		if err := p.SignVote(p.Config.FilePV.ChainID, req.Vote); err != nil {
			p.Logger.Printf("[ERR] pairmint: error while signing %v for height %v: %v\n", req.Vote.Type, req.Vote.Height, err)
			resp.Error = &privvalproto.RemoteSignerError{Description: err.Error()}
		} else {
			p.Logger.Printf("[DEBUG] pairmint: Signed %v for block height %v (signature: %v)\n", req.Vote.Type, req.Vote.Height, strings.ToUpper(hex.EncodeToString(req.Vote.Signature)))
			resp.Vote = *req.Vote
		}
	} else {
		p.Logger.Printf("[DEBUG] pairmint: Validator is ranked #%v, no permission to sign %v for height %v\n", p.Config.Init.Rank, req.Vote.Type, req.Vote.Height)
		resp.Error = &privvalproto.RemoteSignerError{Description: ErrNoSigner.Error()}
	}

	// Send response to Tendermint.
	p.Logger.Printf("[DEBUG] pairmint: Write SignedVoteResponse: %v\n", resp)
	if _, err := rwc.Writer.WriteMsg(wrapMsg(resp)); err != nil {
		return err
	}

	return nil
}

// handleSignProposalRequest handles incoming proposal signing requests.
func (p *PairmintFilePV) handleSignProposalRequest(req *privvalproto.SignProposalRequest, pubkey crypto.PubKey, rwc *connection.ReadWriteConn) error {
	// Prepare empty proposal response.
	resp := &privvalproto.SignedProposalResponse{}

	// Only check the commitsigs once for each block height.
	// Since p.CurrentHeight is initialized to 1, the check for the genesis block is
	// skipped as there is no previous commit to be fetched.
	if req.Proposal.Height > p.CurrentHeight {
		// Retrieve last height's commit from the /commit endpoint of the validator.
		commitsigs, err := connection.GetCommitSigs(req.Proposal.Height - 1)
		if err != nil {
			p.Logger.Printf("[ERR] pairmint: couldn't get commitsigs: %v\n", err)
			resp.Error = &privvalproto.RemoteSignerError{Description: err.Error()}

			// Send error to Tendermint that the commitsigs could not be retrieved.
			// In this case, pairmint can't know whether it is safe to sign or not,
			// so it won't sign the message.
			if _, err := rwc.Writer.WriteMsg(wrapMsg(resp)); err != nil {
				return err
			}

			return ErrNoCommitSigs
		}

		p.Logger.Printf("[DEBUG] pairmint: GET /commit?height=%v: %v\n", req.Proposal.Height-1, commitsigs)
		p.CurrentHeight = req.Proposal.Height

		// Check if the last commit contains our validator's signature.
		if hasSignedCommit(pubkey.Address(), commitsigs) {
			p.Logger.Printf("[DEBUG] pairmint: Found signature from %v in commitsigs from height %v\n", pubkey.Address().String(), req.Proposal.Height-1)
			p.Reset()
		} else {
			p.Logger.Printf("[ERR] pairmint: no commitsig from %v for block height %v\n", pubkey.Address().String(), req.Proposal.Height-1)

			// None of the commitsigs had an entry with our validator's address and
			// a signature in them which means that this block was missed.
			if err := p.Missed(); err != nil {
				p.Logger.Println("[ERR] pairmint: too many missed blocks in a row, updating ranks...")

				// If an error is thrown it means that the threshold of too many missed
				// blocks in a row has been exceeded. Now, a rank update is done in order
				// to replace the signer.
				p.Update()
				p.Reset()
			}
		}
	}

	// After the commitsigs have been checked, check if the validator has permission to sign the proposal.
	if p.Config.Init.Rank == 1 {
		p.Logger.Printf("[DEBUG] pairmint: Validator is ranked #1, signing %v...", req.Proposal.Type)

		// Sign the vote.
		if err := p.SignProposal(p.Config.FilePV.ChainID, req.Proposal); err != nil {
			p.Logger.Printf("[ERR] pairmint: error while signing %v for height %v: %v\n", err, req.Proposal.Type, req.Proposal.Height)
			resp.Error = &privvalproto.RemoteSignerError{Description: err.Error()}
		} else {
			p.Logger.Printf("[DEBUG] pairmint: Signed %v for block height %v (signature: %v)\n", req.Proposal.Type, req.Proposal.Height, strings.ToUpper(hex.EncodeToString(req.Proposal.Signature)))
			resp.Proposal = *req.Proposal
		}
	} else {
		p.Logger.Printf("[DEBUG] pairmint: Validator is ranked #%v, no permission to sign %v for height %v\n", p.Config.Init.Rank, req.Proposal.Type, req.Proposal.Height)
		resp.Error = &privvalproto.RemoteSignerError{Description: ErrNoSigner.Error()}
	}

	// Send response to Tendermint.
	p.Logger.Printf("[DEBUG] pairmint: Write SignedProposalResponse: %v\n", resp)
	if _, err := rwc.Writer.WriteMsg(wrapMsg(resp)); err != nil {
		return err
	}

	return nil
}

// HandleMessage handles all incoming messages from Tendermint.
func (p *PairmintFilePV) HandleMessage(msg *privvalproto.Message, pubkey crypto.PubKey, rwc *connection.ReadWriteConn) error {
	// TODO: Check if requests originate from the chainid specified in the pairmint.toml.

	switch msg.GetSum().(type) {
	case *privvalproto.Message_PingRequest:
		p.Logger.Printf("[DEBUG] pairmint: PingRequest")

		if err := p.handlePingRequest(rwc); err != nil {
			return err
		}

	case *privvalproto.Message_PubKeyRequest:
		req := msg.GetPubKeyRequest()
		p.Logger.Printf("[DEBUG] pairmint: PubKeyRequest: { \"chain_id\": %v }\n", req.ChainId)

		if err := p.handlePubKeyRequest(req, pubkey, rwc); err != nil {
			return err
		}

	case *privvalproto.Message_SignVoteRequest:
		req := msg.GetSignVoteRequest()
		p.Logger.Printf("[DEBUG] pairmint: SignVoteRequest: { \"type\": %v, \"height\": %v, \"round\": %v }\n", req.Vote.Type.String(), req.Vote.Height, req.Vote.Round)

		// TODO: Need to repeat in order to make sure Tendermint gets a response?
		if err := p.handleSignVoteRequest(req, pubkey, rwc); err != nil {
			return err
		}

	case *privvalproto.Message_SignProposalRequest:
		req := msg.GetSignProposalRequest()
		p.Logger.Printf("[DEBUG] pairmint: SignProposalRequest: { \"type\": %v, \"height\": %v, \"round\": %v }\n", req.Proposal.Type.String(), req.Proposal.Height, req.Proposal.Round)

		// TODO: Need to repeat in order to make sure Tendermint gets a response?
		if err := p.handleSignProposalRequest(req, pubkey, rwc); err != nil {
			return err
		}

	default:
		panic(fmt.Errorf("unknown message type: %T", msg.GetSum()))
	}

	return nil
}
