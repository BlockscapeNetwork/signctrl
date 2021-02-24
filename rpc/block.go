package rpc

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

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

// GetBlock gets the block for the specified height.
func GetBlock(rpcladdr string, height int64, logger *log.Logger) (*tm_coretypes.ResultBlock, error) {
	if height < 1 {
		return nil, fmt.Errorf("block height %v does not exist", height)
	}

	logger.Printf("[DEBUG] signctrl: GET /block?height=%v", height)
	laddrWithoutProtocol := strings.SplitAfter(rpcladdr, "://")
	client := &http.Client{Timeout: 5 * time.Second} // TODO: Only timeout if new sign request comes in.
	resp, err := client.Get(fmt.Sprintf("http://%v/block?height=%v", laddrWithoutProtocol[1], height))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	logger.Printf("[DEBUG] signctrl: Received result for GET /block?height=%v", height)

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var block BlockResult
	if err = tm_json.Unmarshal(bytes, &block); err != nil {
		return nil, err
	}
	if block.Result == nil {
		return nil, fmt.Errorf("result block for height %v is nil", height)
	}

	return block.Result, nil
}
