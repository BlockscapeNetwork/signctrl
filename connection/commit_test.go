package connection

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	tmjson "github.com/tendermint/tendermint/libs/json"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"
	tmtypes "github.com/tendermint/tendermint/types"
)

func testCommit() *CommitRPCResponse {
	return &CommitRPCResponse{
		jsonrpc: "2.0",
		id:      1,
		Result: &coretypes.ResultCommit{
			SignedHeader: tmtypes.SignedHeader{
				Commit: &tmtypes.Commit{
					Signatures: []tmtypes.CommitSig{
						{
							ValidatorAddress: []byte("VAL-1-ADDR"),
							Signature:        []byte("VAL-1-SIG"),
						},
						{
							ValidatorAddress: []byte("VAL-2-ADDR"),
							Signature:        []byte("VAL-2-SIG"),
						},
					},
				},
			},
		},
	}
}

func TestGetCommitSigs(t *testing.T) {
	if _, err := GetCommitSigs("", &http.Client{Timeout: 5 * time.Second}, 0); err == nil {
		t.Error("Expected err, instead got nil")
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bytes, err := tmjson.Marshal(testCommit())
		if err != nil {
			t.Fatalf("error while marshaling: %v", err)
		}
		w.Write(bytes)
	}))
	defer ts.Close()

	_, err := GetCommitSigs(strings.TrimPrefix(ts.URL, "http://"), ts.Client(), 1)
	if err != nil {
		t.Errorf("Expected err to be nil, instead got: %v", err)
	}
}
