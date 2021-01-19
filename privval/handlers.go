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
func (p *PairmintFilePV) handlePingRequest(rwc *connection.ReadWriteConn) error {
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
func (p *PairmintFilePV) handleSignVoteRequest(req *privvalproto.SignVoteRequest, pubkey crypto.PubKey, rwc *connection.ReadWriteConn) error {
	// Get commit signatures from the last height.
	commitsigs, err := connection.GetCommitSigs(req.Vote.Height - 1)
	if err != nil {
		return err
	}

	// Prepare empty response.
	resp := &privvalproto.SignedVoteResponse{}

	// Check if the commitsigs have an entry with our validator's address and
	// a signature in it.
	if hasSignedCommit(pubkey.Address().Bytes(), commitsigs) {
		if p.Config.Init.Rank == 1 {
			// Validator is ranked #1, so it has permission to sign the vote.
			if err := p.FilePV.SignVote(p.Config.FilePV.ChainID, req.Vote); err != nil {
				resp.Error.Description = err.Error() // Something went wrong in the signing process.
			} else {
				resp.Vote = *req.Vote // Populate prepared response with signed vote.
			}
		} else {
			// Validator is ranked too low, so it has no signing permission.
			// Reply with a RemoteSignerError.
			resp.Error.Description = ErrNoSigner.Error()
		}
	} else {
		// None of the commitsigs had an entry with our validator's address and
		// a signature in them which means that this block was missed.
		if err := p.Missed(); err != nil {
			// If an error is thrown it means that the threshold of too many missed
			// blocks in a row has been exceeded. Now, a rank update is done in order
			// to replace the signer.
			p.Update()

			// Populate the prepared response with the error.
			resp.Error.Description = err.Error()
		}
	}

	// Send response to Tendermint.
	if _, err := rwc.Writer.WriteMsg(wrapMsg(resp)); err != nil {
		return err
	}

	return nil
}

// handleSignProposalRequest handles incoming proposal signing requests.
func (p *PairmintFilePV) handleSignProposalRequest(req *privvalproto.SignProposalRequest, pubkey crypto.PubKey, rwc *connection.ReadWriteConn) error {
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
func (p *PairmintFilePV) HandleMessage(msg *privvalproto.Message, pubkey crypto.PubKey, rwc *connection.ReadWriteConn) error {
	switch msg.GetSum().(type) {
	case *privvalproto.Message_PingRequest:
		p.Logger.Printf("[DEBUG] pairmint: PingRequest")

		if err := p.handlePingRequest(rwc); err != nil {
			return err
		}

	case *privvalproto.Message_PubKeyRequest:
		req := msg.GetPubKeyRequest()
		p.Logger.Printf("[DEBUG] pairmint: PubKeyRequest for chain ID %v\n", req.ChainId)

		if err := p.handlePubKeyRequest(req, pubkey, rwc); err != nil {
			return err
		}

	case *privvalproto.Message_SignVoteRequest:
		req := msg.GetSignVoteRequest()
		p.Logger.Printf("[DEBUG] pairmint: SignVoteRequest for %v on height %v, round %v\n",
			req.Vote.Type.String(), req.Vote.Height, req.Vote.Round)

		// TODO: Need to repeat in order to make sure Tendermint gets a response?
		if err := p.handleSignVoteRequest(req, pubkey, rwc); err != nil {
			return err
		}

	case *privvalproto.Message_SignProposalRequest:
		req := msg.GetSignProposalRequest()
		p.Logger.Printf("[DEBUG] pairmint: SignProposalRequest for %v on height %v, round %v\n",
			req.Proposal.Type.String(), req.Proposal.Height, req.Proposal.Round)

		if err := p.handleSignProposalRequest(req, pubkey, rwc); err != nil {
			return err
		}

	default:
		panic(fmt.Errorf("unknown message type: %T", msg.GetSum()))
	}

	return nil
}
