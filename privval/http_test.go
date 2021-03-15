package privval

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"testing"

	"github.com/BlockscapeNetwork/signctrl/config"
	"github.com/stretchr/testify/assert"
	tm_privval "github.com/tendermint/tendermint/privval"
)

// getFreePort asks the kernel for a free port that is ready to use.
func getFreePort(t *testing.T) (port int, err error) {
	t.Helper()
	var a *net.TCPAddr
	if a, err = net.ResolveTCPAddr("tcp", "localhost:0"); err == nil {
		var l *net.TCPListener
		if l, err = net.ListenTCP("tcp", a); err == nil {
			defer l.Close()
			return l.Addr().(*net.TCPAddr).Port, nil
		}
	}

	return
}

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
