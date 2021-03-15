package privval

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"testing"

	"github.com/BlockscapeNetwork/signctrl/config"
	"github.com/stretchr/testify/assert"
	tm_privval "github.com/tendermint/tendermint/privval"
)

func mockSCFilePV(t *testing.T) *SCFilePV {
	t.Helper()
	return NewSCFilePV(
		log.New(ioutil.Discard, "", 0),
		config.Config{},
		&config.State{},
		tm_privval.FilePV{},
		&http.Server{Addr: fmt.Sprintf(":%v", DefaultHTTPPort)},
	)
}

func TestGetStatus(t *testing.T) {
	pv := mockSCFilePV(t)
	pv.StartHTTPServer()

	sr, err := GetStatus()
	assert.NotNil(t, sr)
	assert.NoError(t, err)
}
