package privval

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	tm_json "github.com/tendermint/tendermint/libs/json"
)

const (
	// DefaultHTTPPort is the default port which SCFilePV's HTTP server listens on.
	DefaultHTTPPort = 8080
)

// StatusResponse defines the response JSON for status requests.
type StatusResponse struct {
	Height    int64 `json:"height"`
	Rank      int   `json:"rank"`
	SetSize   int   `json:"set_size"`
	Counter   int   `json:"counter"`
	Threshold int   `json:"threshold"`
}

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

func (pv *SCFilePV) statusHandler(rw http.ResponseWriter, r *http.Request) {
	bytes, err := tm_json.Marshal(StatusResponse{
		Height:    pv.GetCurrentHeight(),
		Rank:      pv.GetRank(),
		SetSize:   pv.Config.Base.SetSize,
		Counter:   pv.GetMissedInARow(),
		Threshold: pv.GetThreshold(),
	})
	if err != nil {
		rw.Write(nil)
		return
	}

	rw.Write(bytes)
}

// StartHTTPServer starts an HTTP server.
func (pv *SCFilePV) StartHTTPServer() error {
	pv.Logger.Println("[INFO] signctrl: Starting HTTP server...")

	var errCh chan error
	go func() {
		http.HandleFunc("/status", pv.statusHandler)
		if err := pv.HTTP.ListenAndServe(); err != nil {
			errCh <- err
		}
	}()
	select {
	case <-time.After(100 * time.Millisecond):
		return nil
	case err := <-errCh:
		return err
	}
}
