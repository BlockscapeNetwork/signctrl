package privval

import (
	"fmt"

	"github.com/gogo/protobuf/proto"
	privvalproto "github.com/tendermint/tendermint/proto/tendermint/privval"
)

// WrapPrivvalMsg wraps a protobuf message into a privval proto message.
func WrapPrivvalMsg(pb proto.Message) privvalproto.Message {
	wrapperMsg := privvalproto.Message{}

	switch pb := pb.(type) {
	case *privvalproto.Message:
		wrapperMsg = *pb
	case *privvalproto.PubKeyRequest:
		wrapperMsg.Sum = &privvalproto.Message_PubKeyRequest{PubKeyRequest: pb}
	case *privvalproto.PubKeyResponse:
		wrapperMsg.Sum = &privvalproto.Message_PubKeyResponse{PubKeyResponse: pb}
	case *privvalproto.SignVoteRequest:
		wrapperMsg.Sum = &privvalproto.Message_SignVoteRequest{SignVoteRequest: pb}
	case *privvalproto.SignedVoteResponse:
		wrapperMsg.Sum = &privvalproto.Message_SignedVoteResponse{SignedVoteResponse: pb}
	case *privvalproto.SignedProposalResponse:
		wrapperMsg.Sum = &privvalproto.Message_SignedProposalResponse{SignedProposalResponse: pb}
	case *privvalproto.SignProposalRequest:
		wrapperMsg.Sum = &privvalproto.Message_SignProposalRequest{SignProposalRequest: pb}
	case *privvalproto.PingRequest:
		wrapperMsg.Sum = &privvalproto.Message_PingRequest{PingRequest: pb}
	case *privvalproto.PingResponse:
		wrapperMsg.Sum = &privvalproto.Message_PingResponse{PingResponse: pb}
	default:
		panic(fmt.Errorf("unknown message type %T", pb))
	}

	return wrapperMsg
}
