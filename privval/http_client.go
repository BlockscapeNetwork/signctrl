package privval

import (
	"fmt"
	"io/ioutil"
	"net/http"

	tm_json "github.com/tendermint/tendermint/libs/json"
)

const (
	// DefaultHTTPPort is the default port which SCFilePV's HTTP server listens on.
	DefaultHTTPPort = 8080
)

// GetStatus retrieves the node's status in terms of current height, rank
// and blocks missed in a row.
func GetStatus() (*StatusResponse, error) {
	resp, err := http.DefaultClient.Get(fmt.Sprintf("http://127.0.0.1:%v/status", DefaultHTTPPort))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var sr StatusResponse
	if err := tm_json.Unmarshal(bytes, &sr); err != nil {
		return nil, err
	}

	return &sr, nil
}
