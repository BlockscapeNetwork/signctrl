package privval

import (
	"net/http"
	"time"

	tm_json "github.com/tendermint/tendermint/libs/json"
)

// StatusResponse defines the response JSON for status requests.
type StatusResponse struct {
	Height    int64 `json:"height"`
	Rank      int   `json:"rank"`
	SetSize   int   `json:"set_size"`
	Counter   int   `json:"counter"`
	Threshold int   `json:"threshold"`
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
		_, _ = rw.Write(nil)
		return
	}

	_, _ = rw.Write(bytes)
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
