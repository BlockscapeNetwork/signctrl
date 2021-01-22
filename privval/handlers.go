package privval

import (
	"fmt"

	"github.com/BlockscapeNetwork/pairmint/connection"
	"github.com/gogo/protobuf/proto"
	"github.com/tendermint/tendermint/crypto"
	cryptoproto "github.com/tendermint/tendermint/proto/tendermint/crypto"
	privvalproto "github.com/tendermint/tendermint/proto/tendermint/privval"
	"github.com/tendermint/tendermint/types"
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

	p.Logger.Println("[DEBUG] pairmint: Write PingResponse")

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

	if _, err := rwc.Writer.WriteMsg(wrapMsg(resp)); err != nil {
		return err
	}

	p.Logger.Printf("[DEBUG] pairmint: Write PubKeyResponse: %v\n", resp)

	return nil
}

// handleSignVoteRequest handles incoming vote signing requests.
func (p *PairmintFilePV) handleSignVoteRequest(req *privvalproto.SignVoteRequest, pubkey crypto.PubKey, rwc *connection.ReadWriteConn) error {
	// Prepare empty vote response.
	resp := &privvalproto.SignedVoteResponse{}

	// Check sync info.
	p.Logger.Println("[DEBUG] pairmint: http request: GET /status")
	info, err := connection.GetSyncInfo()
	if err != nil {
		p.Logger.Printf("[ERR] pairmint: couldn't get sync info: %v\n", err)
		return err
	}

	p.Logger.Printf("[DEBUG] pairmint: http response: GET /status: %v\n", info)

	// Check whether the validator is caught up.
	if info.CatchingUp {
		p.Logger.Printf("[INFO] pairmint: Validator is catching up (latest height: %v)\n", info.LatestBlockHeight)

		resp.Error.Description = ErrCatchingUp.Error()
		if _, err := rwc.Writer.WriteMsg(wrapMsg(resp)); err != nil {
			return err
		}

		return ErrCatchingUp
	}

	// Get commit signatures from the last height.
	var commitsigs *[]types.CommitSig
	if req.Vote.Height != 1 {
		p.Logger.Printf("[DEBUG] pairmint: http request: GET /commit?height=%v\n", req.Vote.Height-1)
		commitsigs, err = connection.GetCommitSigs(req.Vote.Height - 1)
		if err != nil {
			return err
		}

		p.Logger.Printf("[DEBUG] pairmint: http response: GET /commit?height=%v: %v\n", req.Vote.Height-1, commitsigs)
	}

	// Check if the commitsigs have an entry with our validator's address and
	// a signature in it.
	if hasSignedCommit(pubkey.Address().Bytes(), commitsigs) {
		p.Logger.Printf("[DEBUG] pairmint: Found signature from %v in commit from height %v.\n", pubkey.Address(), req.Vote.Height)

		if p.Config.Init.Rank == 1 {
			p.Logger.Printf("[DEBUG] pairmint: Validator is ranked #1, sign vote...\n")

			// Validator is ranked #1, so it has permission to sign the vote.
			if err := p.FilePV.SignVote(p.Config.FilePV.ChainID, req.Vote); err != nil {
				p.Logger.Printf("[ERR] pairmint: error while signing vote: %v\n", err)
				resp.Error.Description = err.Error()
			} else {
				p.Logger.Printf("[DEBUG] pairmint: Signed vote: %v\n", req.Vote)
				resp.Vote = *req.Vote // Populate prepared response with signed vote.
			}
		} else {
			p.Logger.Println("[DEBUG] pairmint: Validator has no permission to sign.")

			// Validator is ranked too low, so it has no signing permission.
			// Reply with a RemoteSignerError.
			resp.Error.Description = ErrNoSigner.Error()
		}
	} else {
		p.Logger.Printf("[DEBUG] pairmint: No signature from %v in commit from height %v.\n", pubkey.Address(), req.Vote.Height)

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

	p.Logger.Printf("[DEBUG] pairmint: Write SignedVoteResponse: %v\n", resp)

	return nil
}

// handleSignProposalRequest handles incoming proposal signing requests.
func (p *PairmintFilePV) handleSignProposalRequest(req *privvalproto.SignProposalRequest, pubkey crypto.PubKey, rwc *connection.ReadWriteConn) error {
	// Prepare empty proposal response.
	resp := &privvalproto.SignedProposalResponse{}

	// Check sync info.
	p.Logger.Println("[DEBUG] pairmint: http request: GET /status")
	info, err := connection.GetSyncInfo()
	if err != nil {
		p.Logger.Printf("[ERR] pairmint: couldn't get sync info: %v\n", err)
		return err
	}

	p.Logger.Printf("[DEBUG] pairmint: http response: GET /status: %v\n", info)

	// Check whether the validator is caught up.
	if info.CatchingUp {
		p.Logger.Printf("[INFO] pairmint: Validator is catching up (latest height: %v)\n", info.LatestBlockHeight)

		resp.Error.Description = ErrCatchingUp.Error()
		if _, err := rwc.Writer.WriteMsg(wrapMsg(resp)); err != nil {
			return err
		}

		return ErrCatchingUp
	}

	// Get commit signatures from the last height.
	var commitsigs *[]types.CommitSig
	if req.Proposal.Height != 1 {
		p.Logger.Printf("[DEBUG] pairmint: http request: GET /commit?height=%v\n", req.Proposal.Height-1)
		commitsigs, err = connection.GetCommitSigs(req.Proposal.Height - 1)
		if err != nil {
			return err
		}

		p.Logger.Printf("[DEBUG] pairmint: http response: GET /commit?height=%v: %v\n", req.Proposal.Height-1, commitsigs)
	}

	// Check if the commitsigs have an entry with our validator's address and
	// a signature in it.
	if hasSignedCommit(pubkey.Address().Bytes(), commitsigs) {
		p.Logger.Printf("[DEBUG] pairmint: Found signature from %v in commit from height %v.\n", pubkey.Address(), req.Proposal.Height)

		if p.Config.Init.Rank == 1 {
			p.Logger.Printf("[DEBUG] pairmint: Validator is ranked #1, sign vote...\n")

			// Validator is ranked #1, so it has permission to sign the proposal.
			if err := p.FilePV.SignProposal(p.Config.FilePV.ChainID, req.Proposal); err != nil {
				resp.Error.Description = err.Error() // Something went wrong in the signing process.
			} else {
				resp.Proposal = *req.Proposal // Populate prepared response with signed proposal.
			}
		} else {
			p.Logger.Println("[DEBUG] pairmint: Validator has no permission to sign.")

			// Validator is ranked too low, so it has no signing permission.
			// Reply with a RemoteSignerError.
			resp.Error.Description = ErrNoSigner.Error()
		}
	} else {
		p.Logger.Printf("[DEBUG] pairmint: No signature from %v in commit from height %v.\n", pubkey.Address(), req.Proposal.Height)

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

	if _, err := rwc.Writer.WriteMsg(wrapMsg(resp)); err != nil {
		return err
	}

	p.Logger.Printf("[DEBUG] pairmint: Write SignedProposalResponse: %v\n", resp)

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
		p.Logger.Printf("[DEBUG] pairmint: PubKeyRequest (Chain ID: %v)\n", req.ChainId)

		if err := p.handlePubKeyRequest(req, pubkey, rwc); err != nil {
			return err
		}

	case *privvalproto.Message_SignVoteRequest:
		req := msg.GetSignVoteRequest()
		p.Logger.Printf("[DEBUG] pairmint: SignVoteRequest (Type: %v, Height %v, Round %v)\n",
			req.Vote.Type.String(), req.Vote.Height, req.Vote.Round)

		// TODO: Need to repeat in order to make sure Tendermint gets a response?
		if err := p.handleSignVoteRequest(req, pubkey, rwc); err != nil {
			return err
		}

	case *privvalproto.Message_SignProposalRequest:
		req := msg.GetSignProposalRequest()
		p.Logger.Printf("[DEBUG] pairmint: SignProposalRequest (Type: %v, Height %v, Round %v)\n",
			req.Proposal.Type.String(), req.Proposal.Height, req.Proposal.Round)

		// TODO: Need to repeat in order to make sure Tendermint gets a response?
		if err := p.handleSignProposalRequest(req, pubkey, rwc); err != nil {
			return err
		}

	default:
		panic(fmt.Errorf("unknown message type: %T", msg.GetSum()))
	}

	return nil
}
