package privval

import (
	"encoding/hex"
	"fmt"
	"net/http"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/BlockscapeNetwork/pairmint/connection"
	tmcrypto "github.com/tendermint/tendermint/crypto"
	cryptoenc "github.com/tendermint/tendermint/crypto/encoding"
	cryptoproto "github.com/tendermint/tendermint/proto/tendermint/crypto"
	privvalproto "github.com/tendermint/tendermint/proto/tendermint/privval"
	tmtypes "github.com/tendermint/tendermint/proto/tendermint/types"
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
func (p *PairmintFilePV) handlePubKeyRequest(req *privvalproto.PubKeyRequest, pubkey tmcrypto.PubKey, rwc *connection.ReadWriteConn) error {
	// Check if requests originate from the chainid specified in the pairmint.toml.
	if req.ChainId != p.Config.FilePV.ChainID {
		p.Logger.Printf("[ERR] pairmint: pubkey request is for the wrong chain ID (%v is not %v)", req.ChainId, p.Config.FilePV.ChainID)
		rse := &privvalproto.RemoteSignerError{Description: ErrWrongChainID.Error()}
		if _, err := rwc.Writer.WriteMsg(wrapMsg(&privvalproto.PubKeyResponse{PubKey: cryptoproto.PublicKey{}, Error: rse})); err != nil {
			return err
		}

		return ErrWrongChainID
	}

	encPub, err := cryptoenc.PubKeyToProto(pubkey)
	if err != nil {
		return err
	}

	p.Logger.Printf("[DEBUG] pairmint: Write PubKeyResponse: %v\n", pubkey.Address())
	if _, err := rwc.Writer.WriteMsg(wrapMsg(&privvalproto.PubKeyResponse{PubKey: encPub, Error: nil})); err != nil {
		return err
	}

	return nil
}

// handleSignVoteRequest handles incoming signing requests for prevotes and precommits.
func (p *PairmintFilePV) handleSignVoteRequest(req *privvalproto.SignVoteRequest, pubkey tmcrypto.PubKey, rwc *connection.ReadWriteConn) error {
	// Check if requests originate from the chainid specified in the pairmint.toml.
	if req.ChainId != p.Config.FilePV.ChainID {
		p.Logger.Printf("[ERR] pairmint: sign vote request is for the wrong chain ID (%v is not %v)", req.ChainId, p.Config.FilePV.ChainID)
		rse := &privvalproto.RemoteSignerError{Description: ErrWrongChainID.Error()}
		if _, err := rwc.Writer.WriteMsg(wrapMsg(&privvalproto.SignedVoteResponse{Vote: tmtypes.Vote{}, Error: rse})); err != nil {
			return err
		}

		return ErrWrongChainID
	}

	// Only check the commitsigs once for each block height.
	// The commitsigs are checked based on the commit from height-2. This makes
	// sure that the /commit endpoint actually has all the commit data at its
	// disposal. Taking height-1 resulted in a race condition in Tendermint where
	// sometimes all commit data was there and sometimes some was missing.
	// Al of this means that commit checks are done from block height 3 upwards.
	if req.Vote.Height > p.CurrentHeight && req.Vote.Height > 2 {
		commitsigs, err := connection.GetCommitSigs(p.Config.Init.ValidatorListenAddrRPC, &http.Client{Timeout: 5 * time.Second}, req.Vote.Height-2)
		if err != nil {
			p.Logger.Printf("[ERR] pairmint: couldn't get commitsigs: %v\n", err)
			rse := &privvalproto.RemoteSignerError{Description: err.Error()}

			// Send error to Tendermint that the commitsigs could not be retrieved.
			// In this case, pairmint can't know whether it is safe to sign or not,
			// so it won't sign the message.
			if _, err := rwc.Writer.WriteMsg(wrapMsg(&privvalproto.SignedVoteResponse{Vote: tmtypes.Vote{}, Error: rse})); err != nil {
				return err
			}

			return ErrNoCommitSigs
		}

		p.Logger.Printf("[DEBUG] pairmint: GET /commit?height=%v: %v\n", req.Vote.Height-1, commitsigs)
		p.CurrentHeight = req.Vote.Height

		if !hasSignedCommit(pubkey.Address(), commitsigs) {
			p.Logger.Printf("[ERR] pairmint: no commitsig from %v for block height %v\n", pubkey.Address().String(), req.Vote.Height-1)

			// None of the commitsigs had an entry with our validator's address and
			// a signature in them which means that this block was missed.
			if err := p.Missed(); err != nil {
				// TODO: Remove this if statement when signer rank demotion gets implemented.
				if p.Config.Init.Rank == 1 {
					p.Logger.Printf("[ERR] pairmint: missed %v/%v blocks in a row. Shutting pairmint down...\n", p.Config.Init.Threshold, p.Config.Init.Threshold)

					// Close the secret connection to Tendermint and shut the signer down
					// before it gets a chance to sign the vote.
					rwc.SecretConn.Close()
					os.Exit(int(syscall.SIGTERM))
				}

				p.Logger.Printf("[ERR] pairmint: %v (%v/%v)\n", err, p.MissedInARow, p.Config.Init.Threshold)

				// If an error is thrown it means that the threshold of too many missed
				// blocks in a row has been exceeded. Now, a rank update is done in order
				// to replace the signer.
				p.Update()
			}
		} else {
			p.Logger.Printf("[DEBUG] pairmint: Found commitsig from %v for block height %v\n", pubkey.Address().String(), req.Vote.Height-1)
			p.Reset()
		}
	}

	// Check if the validator has permission to sign the vote.
	if p.Config.Init.Rank > 1 {
		p.Logger.Printf("[DEBUG] pairmint: Validator is ranked #%v, no permission to sign %v for height %v\n", p.Config.Init.Rank, req.Vote.Type, req.Vote.Height)
		rse := &privvalproto.RemoteSignerError{Description: ErrNoSigner.Error()}
		if _, err := rwc.Writer.WriteMsg(wrapMsg(&privvalproto.SignedVoteResponse{Vote: tmtypes.Vote{}, Error: rse})); err != nil {
			return err
		}

		return ErrNoSigner
	}

	p.Logger.Printf("[DEBUG] pairmint: Validator is ranked #%v, signing %v...\n", p.Config.Init.Rank, req.Vote.Type)

	// Sign the vote.
	if err := p.SignVote(p.Config.FilePV.ChainID, req.Vote); err != nil {
		p.Logger.Printf("[ERR] pairmint: error while signing %v for height %v: %v\n", req.Vote.Type, req.Vote.Height, err)
		rse := &privvalproto.RemoteSignerError{Description: err.Error()}
		if _, err := rwc.Writer.WriteMsg(wrapMsg(&privvalproto.SignedVoteResponse{Vote: tmtypes.Vote{}, Error: rse})); err != nil {
			return err
		}

		return err
	}

	p.Logger.Printf("[DEBUG] pairmint: Signed %v for block height %v (signature: %v)\n", req.Vote.Type, req.Vote.Height, strings.ToUpper(hex.EncodeToString(req.Vote.Signature)))
	if _, err := rwc.Writer.WriteMsg(wrapMsg(&privvalproto.SignedVoteResponse{Vote: *req.Vote, Error: nil})); err != nil {
		return err
	}

	return nil
}

// handleSignProposalRequest handles incoming proposal signing requests.
func (p *PairmintFilePV) handleSignProposalRequest(req *privvalproto.SignProposalRequest, pubkey tmcrypto.PubKey, rwc *connection.ReadWriteConn) error {
	// Check if requests originate from the chainid specified in the pairmint.toml.
	if req.ChainId != p.Config.FilePV.ChainID {
		p.Logger.Printf("[ERR] pairmint: sign proposal request is for the wrong chain ID (%v is not %v)", req.ChainId, p.Config.FilePV.ChainID)
		rse := &privvalproto.RemoteSignerError{Description: ErrWrongChainID.Error()}
		if _, err := rwc.Writer.WriteMsg(wrapMsg(&privvalproto.SignedProposalResponse{Proposal: tmtypes.Proposal{}, Error: rse})); err != nil {
			return err
		}

		return ErrWrongChainID
	}

	// Only check the commitsigs once for each block height.
	// The commitsigs are checked based on the commit from height-2. This makes
	// sure that the /commit endpoint actually has all the commit data at its
	// disposal. Taking height-1 resulted in a race condition in Tendermint where
	// sometimes all commit data was there and sometimes some was missing.
	// Al of this means that commit checks are done from block height 3 upwards.
	if req.Proposal.Height > p.CurrentHeight && req.Proposal.Height > 2 {
		commitsigs, err := connection.GetCommitSigs(p.Config.Init.ValidatorListenAddrRPC, &http.Client{Timeout: 5 * time.Second}, req.Proposal.Height-2)
		if err != nil {
			p.Logger.Printf("[ERR] pairmint: couldn't get commitsigs: %v\n", err)
			rse := &privvalproto.RemoteSignerError{Description: err.Error()}

			// Send error to Tendermint that the commitsigs could not be retrieved.
			// In this case, pairmint can't know whether it is safe to sign or not,
			// so it won't sign the message.
			if _, err := rwc.Writer.WriteMsg(wrapMsg(&privvalproto.SignedProposalResponse{Proposal: tmtypes.Proposal{}, Error: rse})); err != nil {
				return err
			}

			return ErrNoCommitSigs
		}

		p.Logger.Printf("[DEBUG] pairmint: GET /commit?height=%v: %v\n", req.Proposal.Height-1, commitsigs)
		p.CurrentHeight = req.Proposal.Height

		// Check if the last commit contains our validator's signature.
		if !hasSignedCommit(pubkey.Address(), commitsigs) {
			p.Logger.Printf("[ERR] pairmint: no commitsig from %v for block height %v\n", pubkey.Address().String(), req.Proposal.Height-1)

			// None of the commitsigs had an entry with our validator's address and
			// a signature in them which means that this block was missed.
			if err := p.Missed(); err != nil {
				// TODO: Remove this if statement when signer rank demotion gets implemented.
				if p.Config.Init.Rank == 1 {
					p.Logger.Printf("[ERR] pairmint: missed %v/%v blocks in a row. Shutting pairmint down...\n", p.Config.Init.Threshold, p.Config.Init.Threshold)

					// Close the secret connection to Tendermint and shut the signer down
					// before it gets a chance to sign the proposal.
					rwc.SecretConn.Close()
					os.Exit(int(syscall.SIGTERM))
				}

				p.Logger.Printf("[ERR] pairmint: %v (%v/%v)\n", err, p.MissedInARow, p.Config.Init.Threshold)

				// If an error is thrown it means that the threshold of too many missed
				// blocks in a row has been exceeded. Now, a rank update is done in order
				// to replace the signer.
				p.Update()
			}
		} else {
			p.Logger.Printf("[DEBUG] pairmint: Found commitsig from %v for block height %v\n", pubkey.Address().String(), req.Proposal.Height-1)
			p.Reset()
		}
	}

	// Check if the validator has permission to sign the proposal.
	if p.Config.Init.Rank > 1 {
		p.Logger.Printf("[DEBUG] pairmint: Validator is ranked #%v, no permission to sign %v for height %v\n", p.Config.Init.Rank, req.Proposal.Type, req.Proposal.Height)
		rse := &privvalproto.RemoteSignerError{Description: ErrNoSigner.Error()}
		if _, err := rwc.Writer.WriteMsg(wrapMsg(&privvalproto.SignedProposalResponse{Proposal: tmtypes.Proposal{}, Error: rse})); err != nil {
			return err
		}

		return ErrNoSigner
	}

	p.Logger.Printf("[DEBUG] pairmint: Validator is ranked #%v, signing %v...", p.Config.Init.Rank, req.Proposal.Type)

	// Sign the proposal.
	if err := p.SignProposal(p.Config.FilePV.ChainID, req.Proposal); err != nil {
		p.Logger.Printf("[ERR] pairmint: error while signing %v for height %v: %v\n", err, req.Proposal.Type, req.Proposal.Height)
		rse := &privvalproto.RemoteSignerError{Description: err.Error()}
		if _, err := rwc.Writer.WriteMsg(wrapMsg(&privvalproto.SignedProposalResponse{Proposal: tmtypes.Proposal{}, Error: rse})); err != nil {
			return err
		}

		return err
	}

	p.Logger.Printf("[DEBUG] pairmint: Signed %v for block height %v (signature: %v)\n", req.Proposal.Type, req.Proposal.Height, strings.ToUpper(hex.EncodeToString(req.Proposal.Signature)))
	if _, err := rwc.Writer.WriteMsg(wrapMsg(&privvalproto.SignedProposalResponse{Proposal: *req.Proposal, Error: nil})); err != nil {
		return err
	}

	return nil
}

// HandleMessage handles all incoming messages from Tendermint.
func (p *PairmintFilePV) HandleMessage(msg *privvalproto.Message, pubkey tmcrypto.PubKey, rwc *connection.ReadWriteConn) error {
	switch msg.GetSum().(type) {
	case *privvalproto.Message_PingRequest:
		p.Logger.Printf("[DEBUG] pairmint: PingRequest")

		// TODO: Need to repeat in order to make sure Tendermint gets a response?
		if err := p.handlePingRequest(rwc); err != nil {
			return err
		}

	case *privvalproto.Message_PubKeyRequest:
		req := msg.GetPubKeyRequest()
		p.Logger.Printf("[DEBUG] pairmint: PubKeyRequest: { \"chain_id\": %v }\n", req.ChainId)

		// TODO: Need to repeat in order to make sure Tendermint gets a response?
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
