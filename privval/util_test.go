package privval

import (
	"testing"

	privvalproto "github.com/tendermint/tendermint/proto/tendermint/privval"
	tmtypes "github.com/tendermint/tendermint/types"
)

func testValidCommitSigs() *[]tmtypes.CommitSig {
	return &[]tmtypes.CommitSig{
		{
			ValidatorAddress: []byte("VAL-1-ADDR"),
			Signature:        []byte("VAL-1-SIG"),
		},
		{
			ValidatorAddress: []byte("VAL-2-ADDR"),
			Signature:        []byte("VAL-2-SIG"),
		},
	}
}

func testInvalidCommitSigs() *[]tmtypes.CommitSig {
	return &[]tmtypes.CommitSig{
		{
			ValidatorAddress: []byte("VAL-1-ADDR"),
			Signature:        nil,
		},
		{
			ValidatorAddress: nil,
			Signature:        []byte("VAL-2-SIG"),
		},
	}
}

func TestHasSignedCommit(t *testing.T) {
	if !hasSignedCommit([]byte("VAL-1-ADDR"), testValidCommitSigs()) {
		t.Errorf("Expected VAL-1 to have a commitsig, instead it doesn't")
	}
	if hasSignedCommit([]byte("VAL-3-ADDR"), testValidCommitSigs()) {
		t.Errorf("Expected VAL-4 to not have a commitsig, instead it does")
	}
	if hasSignedCommit([]byte("VAL-1-ADDR"), testInvalidCommitSigs()) {
		t.Errorf("Expected VAL-1 to not have a complete commitsig, instead it does")
	}
	if hasSignedCommit([]byte("VAL-2-ADDR"), testInvalidCommitSigs()) {
		t.Errorf("Expected VAL-2 to not have a complete commitsig, instead it does")
	}
}

func TestWrapMsg(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("wrapMsg should have panicked")
		}
	}()

	wrapMsg(&privvalproto.Message{})
	wrapMsg(&privvalproto.PingResponse{})
	wrapMsg(&privvalproto.PubKeyResponse{})
	wrapMsg(&privvalproto.SignedVoteResponse{})
	wrapMsg(&privvalproto.SignedProposalResponse{})
	wrapMsg(nil) // panic
}
