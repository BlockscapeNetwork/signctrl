package rpc

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"

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

// ResultChannelResponse defines the data passed into the result channel that is
// retrieved from the /block endpoint.
type ResultChannelResponse struct {
	Result *tm_coretypes.ResultBlock
	Error  error
}

// QueryBlock gets the block for the specified height.
func QueryBlock(ctx context.Context, rpcladdr string, height int64, resultCh chan *ResultChannelResponse, logger *log.Logger) (*tm_coretypes.ResultBlock, error) {
	if height < 1 {
		return nil, fmt.Errorf("block height %v does not exist", height)
	}

	// Cut the protocol from rpcladdr.
	rpcladdrHostPort := regexp.MustCompile(`(tcp|unix)://`).ReplaceAllString(rpcladdr, "")
	url := fmt.Sprintf("http://%v/block?height=%v", rpcladdrHostPort, height)

	go func() {
		// Query the block.
		logger.Printf("[DEBUG] signctrl: GET /block?height=%v", height)
		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			resultCh <- &ResultChannelResponse{nil, err}
			return
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			resultCh <- &ResultChannelResponse{nil, err}
			return
		}
		logger.Printf("[DEBUG] signctrl: Received result for GET /block?height=%v", height)

		// Read from the response body.
		bytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			resultCh <- &ResultChannelResponse{nil, err}
			return
		}

		var block BlockResult
		if err = tm_json.Unmarshal(bytes, &block); err != nil {
			resultCh <- &ResultChannelResponse{nil, err}
			return
		}
		if block.Result == nil {
			resultCh <- &ResultChannelResponse{nil, err}
			return
		}

		resultCh <- &ResultChannelResponse{block.Result, nil}
	}()

	// Wait for the query to return a result or be canceled.
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("request was canceled")
	case rcr := <-resultCh:
		return rcr.Result, rcr.Error
	}
}
