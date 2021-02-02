package privval

import (
	"testing"

	privvalproto "github.com/tendermint/tendermint/proto/tendermint/privval"
	tmtypes "github.com/tendermint/tendermint/types"
)

func TestHasSignedCommit(t *testing.T) {
	commitsigs := []tmtypes.CommitSig{
		{
			ValidatorAddress: []byte("VAL-1-ADDR"),
			Signature:        []byte("VAL-1-SIG"),
		},
		{
			ValidatorAddress: []byte("VAL-2-ADDR"),
			Signature:        []byte("VAL-2-SIG"),
		},
	}

	// Valid case for VAL-1 since it has a complete commitsig.
	if !hasSignedCommit([]byte("VAL-1-ADDR"), &commitsigs) {
		t.Errorf("Expected VAL-1 to have a commitsig, instead it doesn't")
	}

	// Invalid case for VAL-3 since it doesn't have a commitsig.
	if hasSignedCommit([]byte("VAL-3-ADDR"), &commitsigs) {
		t.Errorf("Expected VAL-4 to not have a commitsig, instead it does")
	}

	// Make commitsig of VAL-1 incomplete and therefore invalid.
	commitsigs[0].Signature = nil
	if hasSignedCommit([]byte("VAL-1-ADDR"), &commitsigs) {
		t.Errorf("Expected VAL-1 to not have a complete commitsig, instead it does")
	}

	// Make commitsig of VAL-2 incomplete and therefore invalid.
	commitsigs[1].ValidatorAddress = nil
	if hasSignedCommit([]byte("VAL-2-ADDR"), &commitsigs) {
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
