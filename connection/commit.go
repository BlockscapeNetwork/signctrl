package connection

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/tendermint/tendermint/types"

	"github.com/tendermint/tendermint/libs/json"

	coretypes "github.com/tendermint/tendermint/rpc/core/types"
)

// CommitRPCResponse is the JSONRPC 2.0 response to the commit RPC request.
type CommitRPCResponse struct {
	jsonrpc string
	id      uint64
	Result  *coretypes.ResultCommit `json:"result"`
}

// GetCommitSigs gets the commit signatures of the specified height.
func GetCommitSigs(rpcladdr string, height int64) (*[]types.CommitSig, error) {
	if height < 2 {
		return nil, fmt.Errorf("can't get commitsigs for block height %v", height)
	}

	client := http.Client{Timeout: 5 * time.Second} // TODO: Timeouts need to be set according to the block time of the chain.
	resp, err := client.Get(fmt.Sprintf("http://%v/commit?height=%v", rpcladdr, height))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var rpc CommitRPCResponse
	if err = json.Unmarshal(bytes, &rpc); err != nil {
		return nil, err
	}

	if rpc.Result == nil {
		return nil, ErrNoCommitSigs
	}

	return &rpc.Result.Commit.Signatures, nil
}
