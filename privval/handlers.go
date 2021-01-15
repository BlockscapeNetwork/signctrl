package privval

import (
	"fmt"

	"github.com/BlockscapeNetwork/pairmint/connection"
	"github.com/gogo/protobuf/proto"
	"github.com/tendermint/tendermint/crypto"
	cryptoproto "github.com/tendermint/tendermint/proto/tendermint/crypto"
	privvalproto "github.com/tendermint/tendermint/proto/tendermint/privval"
)

// wrapMsg wraps a protobuf message into a privval proto message.
func wrapMsg(pb proto.Message) *privvalproto.Message {
	msg := privvalproto.Message{}

	switch pb := pb.(type) {
	case *privvalproto.Message:
		msg = *pb
	case *privvalproto.PingResponse:
		msg.Sum = &privvalproto.Message_PingResponse{PingResponse: pb}
	case *privvalproto.PubKeyResponse:
		msg.Sum = &privvalproto.Message_PubKeyResponse{PubKeyResponse: pb}
	case *privvalproto.SignedVoteResponse:
		msg.Sum = &privvalproto.Message_SignedVoteResponse{SignedVoteResponse: pb}
	case *privvalproto.SignedProposalResponse:
		msg.Sum = &privvalproto.Message_SignedProposalResponse{SignedProposalResponse: pb}
	default:
		panic(fmt.Errorf("unknown message type: %T", pb))
	}

	return &msg
}

// handlePingRequest handles incoming ping requests.
func (p *PairmintFilePV) handlePingRequest(req *privvalproto.PingRequest, rwc *connection.ReadWriteConn) error {
	if _, err := rwc.Writer.WriteMsg(wrapMsg(&privvalproto.PingResponse{})); err != nil {
		return err
	}

	return nil
}

// handlePubKeyRequest handles incoming public key requests.
func (p *PairmintFilePV) handlePubKeyRequest(req *privvalproto.PubKeyRequest, pubkey crypto.PubKey, rwc *connection.ReadWriteConn) error {
	if _, err := rwc.Writer.WriteMsg(wrapMsg(&privvalproto.PubKeyResponse{
		PubKey: cryptoproto.PublicKey{
			Sum: &cryptoproto.PublicKey_Ed25519{
				Ed25519: pubkey.Bytes(),
			},
		}})); err != nil {
		return err
	}

	return nil
}

// handleSignVoteRequest handles incoming vote signing requests.
func (p *PairmintFilePV) handleSignVoteRequest(req *privvalproto.SignVoteRequest, rwc *connection.ReadWriteConn) error {
	// TODO: Sign vote if node is primary, and reply with signed vote.
	// TODO: Else, reply with an error.

	if err := p.FilePV.SignVote(p.Config.FilePV.ChainID, req.Vote); err != nil {
		return err
	}
	if _, err := rwc.Writer.WriteMsg(wrapMsg(&privvalproto.SignedVoteResponse{Vote: *req.Vote})); err != nil {
		return err
	}

	return nil
}

// handleSignProposalRequest handles incoming proposal signing requests.
func (p *PairmintFilePV) handleSignProposalRequest(req *privvalproto.SignProposalRequest, rwc *connection.ReadWriteConn) error {
	// TODO: Sign proposal if node is primary, and reply with signed proposal.
	// TODO: Else, reply with an error.

	if err := p.FilePV.SignProposal(p.Config.FilePV.ChainID, req.Proposal); err != nil {
		return err
	}
	if _, err := rwc.Writer.WriteMsg(wrapMsg(&privvalproto.SignedProposalResponse{Proposal: *req.Proposal})); err != nil {
		return err
	}

	return nil
}

// HandleMessage handles all incoming messages from Tendermint.
func (p *PairmintFilePV) HandleMessage(msg *privvalproto.Message, rwc *connection.ReadWriteConn) error {
	switch msg.GetSum().(type) {
	case *privvalproto.Message_PingRequest:
		p.handlePingRequest(msg.GetPingRequest(), rwc)

	case *privvalproto.Message_PubKeyRequest:
		pubkey, err := p.GetPubKey()
		if err != nil {
			return ErrMissingPubKey
		}
		p.handlePubKeyRequest(msg.GetPubKeyRequest(), pubkey, rwc)

	case *privvalproto.Message_SignVoteRequest:
		p.handleSignVoteRequest(msg.GetSignVoteRequest(), rwc)

	case *privvalproto.Message_SignProposalRequest:
		p.handleSignProposalRequest(msg.GetSignProposalRequest(), rwc)

	default:
		panic(fmt.Errorf("unknown message type: %T", msg.GetSum()))
	}

	return nil
}
