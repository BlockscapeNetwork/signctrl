package rpc

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"

	"github.com/BlockscapeNetwork/signctrl/types"
	tm_json "github.com/tendermint/tendermint/libs/json"
	tm_coretypes "github.com/tendermint/tendermint/rpc/core/types"
)

// BlockResult defines the JSONRPC 2.0 response structure for Tendermint's /block
// endpoint.
type BlockResult struct {
	jsonrpc string
	id      uint64
	Result  *tm_coretypes.ResultBlock `json:"result"`
}

// resultChannelResponse defines the data passed into the result channel that is
// retrieved from the /block endpoint.
type resultChannelResponse struct {
	result *tm_coretypes.ResultBlock
	err    error
}

// QueryBlock gets the block for the specified height.
func QueryBlock(ctx context.Context, rpcladdr string, height int64, logger *types.SyncLogger) (*tm_coretypes.ResultBlock, error) {
	if height < 1 {
		return nil, fmt.Errorf("block height %v does not exist", height)
	}

	// Cut the protocol from rpcladdr.
	rpcladdrHostPort := regexp.MustCompile(`(tcp|unix)://`).ReplaceAllString(rpcladdr, "")
	url := fmt.Sprintf("http://%v/block?height=%v", rpcladdrHostPort, height)
	resultCh := make(chan *resultChannelResponse)

	go func() {
		// Query the block.
		logger.Debug("GET %v", url)
		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			resultCh <- &resultChannelResponse{nil, err}
			return
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			resultCh <- &resultChannelResponse{nil, err}
			return
		}

		// Read from the response body.
		bytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			resultCh <- &resultChannelResponse{nil, err}
			return
		}

		var block BlockResult
		if err := tm_json.Unmarshal(bytes, &block); err != nil {
			resultCh <- &resultChannelResponse{nil, err}
			return
		}
		if block.Result == nil {
			resultCh <- &resultChannelResponse{nil, err}
			return
		}

		resultCh <- &resultChannelResponse{block.Result, nil}
	}()

	// Wait for the query to return a result or be canceled.
	select {
	case <-ctx.Done():
		logger.Debug("Canceled GET %v", url)
		return nil, fmt.Errorf("request was canceled")
	case rcr := <-resultCh:
		logger.Debug("Received result for GET %v", url)
		return rcr.result, rcr.err
	}
}
