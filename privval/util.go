package privval

import (
	"bytes"
	"fmt"

	"github.com/gogo/protobuf/proto"
	privvalproto "github.com/tendermint/tendermint/proto/tendermint/privval"
	tmtypes "github.com/tendermint/tendermint/types"
)

// hasSignedCommit checks whether the given validator address has a commitsig
// in the provided commitsigs.
func hasSignedCommit(valaddr tmtypes.Address, commitsigs *[]tmtypes.CommitSig) bool {
	for _, commitsig := range *commitsigs {
		if cmp := bytes.Compare(commitsig.ValidatorAddress, valaddr); cmp == 0 && commitsig.Signature != nil {
			return true
		}
	}

	return false
}

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
