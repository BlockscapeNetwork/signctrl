package privval

import (
	"io/ioutil"
	"net/http"

	tm_json "github.com/tendermint/tendermint/libs/json"
)

// GetStatus retrieves the node's status in terms of current height, rank
// and blocks missed in a row.
func GetStatus() (*StatusResponse, error) {
	resp, err := http.DefaultClient.Get("http://127.0.0.1:8080/status")
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
