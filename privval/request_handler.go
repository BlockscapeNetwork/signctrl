package privval

import (
	"bytes"
	"context"
	"errors"
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

var (
	// ErrRankObsolete is returned if the requested vote height is too far ahead of the last
	// block the validator signed. The gap must be at least {threshold} blocks.
	ErrRankObsolete = errors.New("at least one threshold was exceeded between requested vote height and last_signed_height")
)

// wrapMsg wraps a protobuf message into a privval proto message.
func wrapMsg(pb proto.Message) *tm_privvalproto.Message {
	msg := tm_privvalproto.Message{}
	switch pb := pb.(type) {
	case *tm_privvalproto.Message:
		msg = *pb
	case *tm_privvalproto.PingRequest:
		msg.Sum = &tm_privvalproto.Message_PingRequest{PingRequest: pb}
	case *tm_privvalproto.PingResponse:
		msg.Sum = &tm_privvalproto.Message_PingResponse{PingResponse: pb}
	case *tm_privvalproto.PubKeyRequest:
		msg.Sum = &tm_privvalproto.Message_PubKeyRequest{PubKeyRequest: pb}
	case *tm_privvalproto.PubKeyResponse:
		msg.Sum = &tm_privvalproto.Message_PubKeyResponse{PubKeyResponse: pb}
	case *tm_privvalproto.SignVoteRequest:
		msg.Sum = &tm_privvalproto.Message_SignVoteRequest{SignVoteRequest: pb}
	case *tm_privvalproto.SignedVoteResponse:
		msg.Sum = &tm_privvalproto.Message_SignedVoteResponse{SignedVoteResponse: pb}
	case *tm_privvalproto.SignProposalRequest:
		msg.Sum = &tm_privvalproto.Message_SignProposalRequest{SignProposalRequest: pb}
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

// isRankUpToDate checks whether the validator's rank is still up to date or obsolete.
func isRankUpToDate(reqHeight int64, lastHeight int64, threshold int) bool {
	return reqHeight-lastHeight < int64(threshold+1)
}

// handlePingRequest handles a PingRequest by returning a
// PingResponse.
func handlePingRequest(pv *SCFilePV) (*tm_privvalproto.Message, error) {
	pv.Logger.Debug("Received PingRequest")
	return wrapMsg(&tm_privvalproto.PingResponse{}), nil
}

// handlePubKeyRequest handles a PubKeyRequest by returning a
// PubKeyResponse.
func handlePubKeyRequest(req *tm_privvalproto.PubKeyRequest, pv *SCFilePV) (*tm_privvalproto.Message, error) {
	pv.Logger.Debug("Received PubKeyRequest: %v", req) // TODO: Add toString() for tm_privvalproto.PubKeyRequest

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

// sharedSignRequestData defines data shared between votes and proposals.
type sharedSignRequestData struct {
	chainID string
	msgType tm_typesproto.SignedMsgType
	height  int64
}

// getSharedSignRequestData returns shared sign request data.
func getSharedSignRequestData(msg *tm_privvalproto.Message) (data sharedSignRequestData) {
	switch msg.Sum.(type) {
	case *tm_privvalproto.Message_SignVoteRequest:
		req := msg.GetSignVoteRequest()
		data.chainID = req.ChainId
		data.msgType = req.Vote.Type
		data.height = req.Vote.Height

	case *tm_privvalproto.Message_SignProposalRequest:
		req := msg.GetSignProposalRequest()
		data.chainID = req.ChainId
		data.msgType = req.Proposal.Type
		data.height = req.Proposal.Height
	}

	return data
}

// buildResponse builds a response for the given message. The message must wrap
// either a SignVoteRequest or a SignProposalRequest.
func buildResponse(msg *tm_privvalproto.Message, rse *tm_privvalproto.RemoteSignerError) *tm_privvalproto.Message {
	switch msg.Sum.(type) {
	case *tm_privvalproto.Message_SignVoteRequest:
		return wrapMsg(&tm_privvalproto.SignedVoteResponse{
			Vote:  *msg.GetSignVoteRequest().GetVote(),
			Error: rse,
		})

	case *tm_privvalproto.Message_SignProposalRequest:
		return wrapMsg(&tm_privvalproto.SignedProposalResponse{
			Proposal: *msg.GetSignProposalRequest().GetProposal(),
			Error:    rse,
		})
	}

	return nil
}

// handleSignRequest handles SignVoteRequests and SignProposalRequests by
// returning either a SignedVoteResponse or a SignedProposalResponse.
func handleSignRequest(ctx context.Context, msg *tm_privvalproto.Message, pv *SCFilePV) (*tm_privvalproto.Message, error) {
	reqData := getSharedSignRequestData(msg)

	// Check if the request is for the chain ID specified in the config.toml.
	if reqData.chainID != pv.Config.Privval.ChainID {
		err := fmt.Errorf("expected sign request for chain ID '%v', instead got '%v'", pv.Config.Privval.ChainID, reqData.chainID)
		return buildResponse(msg, &tm_privvalproto.RemoteSignerError{Description: err.Error()}), err
	}

	// If the requested height is at least {threshold}+1 higher than last_signed_height,
	// the node's rank has become obsolete due to a rank update in the set.
	if !isRankUpToDate(reqData.height, pv.State.LastHeight, pv.GetThreshold()) {
		pv.Logger.Debug("The requested height differs too much from the last height (%v - %v >= %v)", reqData.height, pv.State.LastHeight, pv.GetThreshold()+1)
		return buildResponse(msg, &tm_privvalproto.RemoteSignerError{Description: ErrRankObsolete.Error()}), ErrRankObsolete
	}

	// Only check the commitsigs once for each block height.
	// Also, only start checking for block heights greater than 1.
	// This is due to the genesis block not having any commitsigs.
	if reqData.height > pv.BaseSignCtrled.GetCurrentHeight() && reqData.height > 1 {
		// Get block information from the validator's /block endpoint.
		rb, err := rpc.QueryBlock(ctx, pv.Config.Base.ValidatorListenAddressRPC, reqData.height-1, pv.Logger)
		if err != nil {
			return buildResponse(msg, &tm_privvalproto.RemoteSignerError{Description: err.Error()}), err
		}

		// Update the current height to the height of the request.
		pv.BaseSignCtrled.SetCurrentHeight(reqData.height)
		pv.State.LastHeight = reqData.height

		// Check if the commitsigs in the block are signed by the validator.
		pub, _ := pv.TMFilePV.GetPubKey()
		if !hasSignedCommit(pub.Address(), &rb.Block.LastCommit.Signatures) {
			// Check if the threshold of too many missed blocks in a row is exceeded.
			if err := pv.Missed(); err != nil {
				if err == types.ErrMustShutdown {
					return buildResponse(msg, &tm_privvalproto.RemoteSignerError{Description: err.Error()}), err
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
		err := fmt.Errorf("no signing permission for %v on block height %v (rank: %v)", reqData.msgType, reqData.height, pv.GetRank())
		return buildResponse(msg, &tm_privvalproto.RemoteSignerError{Description: err.Error()}), err
	}

	switch msg.Sum.(type) {
	case *tm_privvalproto.Message_SignVoteRequest:
		req := msg.GetSignVoteRequest()

		// The node has permission to sign the vote, so sign it.
		if err := pv.TMFilePV.SignVote(pv.Config.Privval.ChainID, req.Vote); err != nil {
			err := fmt.Errorf("failed to sign %v for block height %v: %v", req.Vote.Type, req.Vote.Height, err)
			return buildResponse(msg, &tm_privvalproto.RemoteSignerError{Description: err.Error()}), err
		}

		pv.Logger.Info("Signed %v for block height %v", req.Vote.Type, req.Vote.Height)
		return buildResponse(wrapMsg(&tm_privvalproto.SignVoteRequest{Vote: req.Vote, ChainId: req.GetChainId()}), nil), nil

	case *tm_privvalproto.Message_SignProposalRequest:
		req := msg.GetSignProposalRequest()

		// The node has permission to sign the proposal, so sign it.
		if err := pv.TMFilePV.SignProposal(pv.Config.Privval.ChainID, req.Proposal); err != nil {
			err := fmt.Errorf("failed to sign %v for block height %v: %v", req.Proposal.Type, req.Proposal.Height, err)
			return buildResponse(msg, &tm_privvalproto.RemoteSignerError{Description: err.Error()}), err
		}

		pv.Logger.Info("Signed %v for block height %v", req.Proposal.Type, req.Proposal.Height)
		return buildResponse(wrapMsg(&tm_privvalproto.SignProposalRequest{Proposal: req.Proposal, ChainId: req.GetChainId()}), nil), nil

	default:
		return nil, fmt.Errorf("unknown sign request: %T", msg)
	}
}

// HandleRequest handles all incoming requests from the validator.
func HandleRequest(ctx context.Context, msg *tm_privvalproto.Message, pv *SCFilePV) (*tm_privvalproto.Message, error) {
	switch msg.Sum.(type) {
	case *tm_privvalproto.Message_PingRequest:
		return handlePingRequest(pv)
	case *tm_privvalproto.Message_PubKeyRequest:
		return handlePubKeyRequest(msg.GetPubKeyRequest(), pv)
	case *tm_privvalproto.Message_SignVoteRequest, *tm_privvalproto.Message_SignProposalRequest:
		return handleSignRequest(ctx, msg, pv)
	default:
		return nil, fmt.Errorf("unknown message: %v", msg)
	}
}
