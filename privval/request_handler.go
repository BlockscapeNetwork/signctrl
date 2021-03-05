package privval

import (
	"bytes"
	"context"
	"fmt"

	"github.com/BlockscapeNetwork/signctrl/rpc"
	"github.com/BlockscapeNetwork/signctrl/types"
	"github.com/gogo/protobuf/proto"
	tm_cryptoenc "github.com/tendermint/tendermint/crypto/encoding"
	tm_cryptoproto "github.com/tendermint/tendermint/proto/tendermint/crypto"
	tm_privvalproto "github.com/tendermint/tendermint/proto/tendermint/privval"
	tm_typesproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tm_types "github.com/tendermint/tendermint/types"
)

// wrapMsg wraps a protobuf message into a privval proto message.
func wrapMsg(pb proto.Message) *tm_privvalproto.Message {
	msg := tm_privvalproto.Message{}
	switch pb := pb.(type) {
	case *tm_privvalproto.Message:
		msg = *pb
	case *tm_privvalproto.PingResponse:
		msg.Sum = &tm_privvalproto.Message_PingResponse{PingResponse: pb}
	case *tm_privvalproto.PubKeyResponse:
		msg.Sum = &tm_privvalproto.Message_PubKeyResponse{PubKeyResponse: pb}
	case *tm_privvalproto.SignedVoteResponse:
		msg.Sum = &tm_privvalproto.Message_SignedVoteResponse{SignedVoteResponse: pb}
	case *tm_privvalproto.SignedProposalResponse:
		msg.Sum = &tm_privvalproto.Message_SignedProposalResponse{SignedProposalResponse: pb}
	default:
		panic(fmt.Errorf("unknown message type: %T", pb))
	}

	return &msg
}

// hasSignedCommit checks whether the given validator address has a commitsig
// in the provided commitsigs.
func hasSignedCommit(valaddr tm_types.Address, commitsigs *[]tm_types.CommitSig) bool {
	for _, commitsig := range *commitsigs {
		if cmp := bytes.Compare(commitsig.ValidatorAddress, valaddr); cmp == 0 {
			return true
		}
	}

	return false
}

// handlePingRequest handles a PingRequest by returning a
// PingResponse.
func handlePingRequest(pv *SCFilePV) (*tm_privvalproto.Message, error) {
	pv.Logger.Println("[DEBUG] signctrl: Received PingRequest")
	return wrapMsg(&tm_privvalproto.PingResponse{}), nil
}

// handlePubKeyRequest handles a PubKeyRequest by returning a
// PubKeyResponse.
func handlePubKeyRequest(req *tm_privvalproto.PubKeyRequest, pv *SCFilePV) (*tm_privvalproto.Message, error) {
	pv.Logger.Printf("[DEBUG] signctrl: Received PubKeyRequest: %v", req) // TODO: Add toString() for tm_privvalproto.PubKeyRequest

	// Check if the PubKeyRequest is for the chain ID specified
	// in the config.toml.
	if req.GetChainId() != pv.Config.Privval.ChainID {
		err := fmt.Errorf("expected PubKeyRequest for chain ID '%v', instead got '%v'", pv.Config.Privval.ChainID, req.GetChainId())
		return wrapMsg(&tm_privvalproto.PubKeyResponse{
			PubKey: tm_cryptoproto.PublicKey{},
			Error:  &tm_privvalproto.RemoteSignerError{Description: err.Error()},
		}), err
	}

	// GetPubKey never returns an error since the pubkey returned
	// is never nil.
	pubkey, _ := pv.TMFilePV.GetPubKey()
	pbEncPub, err := tm_cryptoenc.PubKeyToProto(pubkey)
	if err != nil {
		return wrapMsg(&tm_privvalproto.PubKeyResponse{
			PubKey: tm_cryptoproto.PublicKey{},
			Error:  &tm_privvalproto.RemoteSignerError{Description: err.Error()},
		}), err
	}

	return wrapMsg(&tm_privvalproto.PubKeyResponse{
		PubKey: pbEncPub,
		Error:  nil,
	}), nil
}

// handleSignVoteRequest handles a SignVoteRequest by returning
// a SignVoteResponse.
func handleSignVoteRequest(ctx context.Context, req *tm_privvalproto.SignVoteRequest, pv *SCFilePV) (*tm_privvalproto.Message, error) {
	pv.Logger.Printf("[DEBUG] signctrl: Received SignVoteRequest: %v", req) // TODO: Add toString() for tm_privvalproto.SignVoteRequest

	// Check if the SignVoteRequest is for the chain ID specified
	// in the config.toml.
	if req.GetChainId() != pv.Config.Privval.ChainID {
		err := fmt.Errorf("expected SignVoteRequest for chain ID '%v', instead got '%v'", pv.Config.Privval.ChainID, req.GetChainId())
		return wrapMsg(&tm_privvalproto.SignedVoteResponse{
			Vote:  tm_typesproto.Vote{},
			Error: &tm_privvalproto.RemoteSignerError{Description: err.Error()},
		}), err
	}

	// Only check the commitsigs once for each block height.
	// Also, only start checking for block heights greater than 1.
	// This is due to the genesis block not having any commitsigs.
	if req.Vote.Height > pv.BaseSignCtrled.GetCurrentHeight() && req.Vote.Height > 1 {
		// Get block information from the validator's /block endpoint.
		rb, err := rpc.QueryBlock(ctx, pv.Config.Base.ValidatorListenAddressRPC, req.Vote.Height-1, pv.Logger)
		if err != nil {
			return wrapMsg(&tm_privvalproto.SignedVoteResponse{
				Vote:  tm_typesproto.Vote{},
				Error: &tm_privvalproto.RemoteSignerError{Description: err.Error()},
			}), err
		}

		// Update the current height to the height of the request.
		pv.BaseSignCtrled.SetCurrentHeight(req.Vote.Height)

		// Check if the commitsigs in the block are signed by the validator.
		if !hasSignedCommit(pv.TMFilePV.GetAddress(), &rb.Block.LastCommit.Signatures) {
			// Check if the threshold of too many missed blocks in a row is exceeded.
			if err := pv.Missed(); err != nil {
				if err == types.ErrMustShutdown {
					return wrapMsg(&tm_privvalproto.SignedVoteResponse{
						Vote:  tm_typesproto.Vote{},
						Error: &tm_privvalproto.RemoteSignerError{Description: err.Error()},
					}), err
				}
			}
		} else {
			// If the commit was signed, reset the counter for missed blocks in a row
			// and unlock it if it hasn't already been unlocked.
			pv.Reset()
			pv.UnlockCounter()
		}
	}

	// Prevent the node from signing if it's not ranked first in the set.
	if pv.GetRank() > 1 {
		err := fmt.Errorf("no signing permission for %v on block height %v (rank: %v)", req.Vote.Type, req.Vote.Height, pv.GetRank())
		return wrapMsg(&tm_privvalproto.SignedVoteResponse{
			Vote:  tm_typesproto.Vote{},
			Error: &tm_privvalproto.RemoteSignerError{Description: err.Error()},
		}), err
	}

	// The node has permission to sign the vote, so sign it.
	if err := pv.TMFilePV.SignVote(pv.Config.Privval.ChainID, req.Vote); err != nil {
		err := fmt.Errorf("failed to sign %v for block height %v: %v", req.Vote.Type, req.Vote.Height, err)
		return wrapMsg(&tm_privvalproto.SignedVoteResponse{
			Vote:  tm_typesproto.Vote{},
			Error: &tm_privvalproto.RemoteSignerError{Description: err.Error()},
		}), err
	}

	pv.Logger.Printf("[INFO] signctrl: Signed %v for block height %v", req.Vote.Type, req.Vote.Height)
	return wrapMsg(&tm_privvalproto.SignedVoteResponse{
		Vote:  *req.Vote,
		Error: nil,
	}), nil
}

// handleSignProposalRequest handles a SignProposalRequest by
// returning a SignProposalResponse.
func handleSignProposalRequest(ctx context.Context, req *tm_privvalproto.SignProposalRequest, pv *SCFilePV) (*tm_privvalproto.Message, error) {
	pv.Logger.Printf("[DEBUG] signctrl: Received SignProposalRequest: %v", req) // TODO: Add toString() for tm_privvalproto.SignProposalRequest

	// Check if the SignProposalRequest is for the chain ID specified
	// in the config.toml.
	if req.GetChainId() != pv.Config.Privval.ChainID {
		err := fmt.Errorf("expected SignProposalRequest for chain ID '%v', instead got '%v'", pv.Config.Privval.ChainID, req.GetChainId())
		return wrapMsg(&tm_privvalproto.SignedProposalResponse{
			Proposal: tm_typesproto.Proposal{},
			Error:    &tm_privvalproto.RemoteSignerError{Description: err.Error()},
		}), err
	}

	// Only check the commitsigs once for each block height.
	// Also, only start checking for block heights greater than 1.
	// This is due to the genesis block not having any commitsigs.
	if req.Proposal.Height > pv.BaseSignCtrled.GetCurrentHeight() && req.Proposal.Height > 1 {
		// Get block information from the validator's /block endpoint.
		rb, err := rpc.QueryBlock(ctx, pv.Config.Base.ValidatorListenAddressRPC, req.Proposal.Height-1, pv.Logger)
		if err != nil {
			return wrapMsg(&tm_privvalproto.SignedProposalResponse{
				Proposal: tm_typesproto.Proposal{},
				Error:    &tm_privvalproto.RemoteSignerError{Description: err.Error()},
			}), err
		}

		// Update the current height to the height of the request.
		pv.BaseSignCtrled.SetCurrentHeight(req.Proposal.Height)

		// Check if the commitsigs in the block are signed by the validator.
		if !hasSignedCommit(pv.TMFilePV.GetAddress(), &rb.Block.LastCommit.Signatures) {
			// Check if the threshold of too many missed blocks in a row is exceeded.
			if err := pv.Missed(); err != nil {
				if err == types.ErrMustShutdown {
					return wrapMsg(&tm_privvalproto.SignedProposalResponse{
						Proposal: tm_typesproto.Proposal{},
						Error:    &tm_privvalproto.RemoteSignerError{Description: err.Error()},
					}), err
				}
			}
		} else {
			// If the commit was signed, reset the counter for missed blocks in a row
			// and unlock it if it hasn't already been unlocked.
			pv.Reset()
			pv.UnlockCounter()
		}
	}

	// Prevent the node from signing if it's not ranked first in the set.
	if pv.GetRank() > 1 {
		err := fmt.Errorf("no signing permission for %v on block height %v (rank: %v)", req.Proposal.Type, req.Proposal.Height, pv.GetRank())
		return wrapMsg(&tm_privvalproto.SignedProposalResponse{
			Proposal: tm_typesproto.Proposal{},
			Error:    &tm_privvalproto.RemoteSignerError{Description: err.Error()},
		}), err
	}

	// The node has permission to sign the proposal, so sign it.
	if err := pv.TMFilePV.SignProposal(pv.Config.Privval.ChainID, req.Proposal); err != nil {
		err := fmt.Errorf("failed to sign %v for block height %v: %v", req.Proposal.Type, req.Proposal.Height, err)
		return wrapMsg(&tm_privvalproto.SignedProposalResponse{
			Proposal: tm_typesproto.Proposal{},
			Error:    &tm_privvalproto.RemoteSignerError{Description: err.Error()},
		}), err
	}

	pv.Logger.Printf("[INFO] signctrl: Signed %v for block height %v", req.Proposal.Type, req.Proposal.Height)
	return wrapMsg(&tm_privvalproto.SignedProposalResponse{
		Proposal: *req.Proposal,
		Error:    nil,
	}), nil
}

// HandleRequest handles all incoming requests from the validator.
func HandleRequest(ctx context.Context, msg *tm_privvalproto.Message, pv *SCFilePV) (*tm_privvalproto.Message, error) {
	switch msg.Sum.(type) {
	case *tm_privvalproto.Message_PingRequest:
		return handlePingRequest(pv)
	case *tm_privvalproto.Message_PubKeyRequest:
		return handlePubKeyRequest(msg.GetPubKeyRequest(), pv)
	case *tm_privvalproto.Message_SignVoteRequest:
		return handleSignVoteRequest(ctx, msg.GetSignVoteRequest(), pv)
	case *tm_privvalproto.Message_SignProposalRequest:
		return handleSignProposalRequest(ctx, msg.GetSignProposalRequest(), pv)
	default:
		return nil, fmt.Errorf("unknown message: %v", msg)
	}
}
