package connection

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/tendermint/tendermint/libs/json"

	coretypes "github.com/tendermint/tendermint/rpc/core/types"
)

// StatusRPCResponse is the JSONRPC 2.0 response to the status RPC request.
type StatusRPCResponse struct {
	jsonrpc string
	id      uint64
	Result  *coretypes.ResultStatus `json:"result"`
}

// GetSyncInfo gets the current sync info from the Tendermint validator.
func GetSyncInfo() (*coretypes.SyncInfo, error) {
	client := http.Client{Timeout: 10 * time.Second}
	url := fmt.Sprintf("http://127.0.0.1:26657/status") // TODO: Replace hardcoded address with config address ([rpc].laddr)

	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var rpc StatusRPCResponse
	if err = json.Unmarshal(bytes, &rpc); err != nil {
		return nil, err
	}

	return &rpc.Result.SyncInfo, nil
}
