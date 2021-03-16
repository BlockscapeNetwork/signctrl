package privval

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"testing"

	"github.com/BlockscapeNetwork/signctrl/config"
	tm_ed25519 "github.com/tendermint/tendermint/crypto/ed25519"
	tm_privval "github.com/tendermint/tendermint/privval"
)

func testConfig(t *testing.T) config.Config {
	t.Helper()
	return config.Config{
		Base: config.Base{
			LogLevel:                  "INFO",
			SetSize:                   2,
			Threshold:                 10,
			StartRank:                 1,
			ValidatorListenAddress:    "tcp://127.0.0.1:3000",
			ValidatorListenAddressRPC: "tcp://127.0.0.1:26657",
			RetryDialAfter:            "15s",
		},
		Privval: config.PrivValidator{
			ChainID: "testchain",
		},
	}
}

func testState(t *testing.T) *config.State {
	t.Helper()
	return &config.State{
		LastHeight: 1,
		LastRank:   1,
	}
}

func testFilePV(t *testing.T) tm_privval.FilePV {
	t.Helper()
	priv := tm_ed25519.GenPrivKey()
	return tm_privval.FilePV{
		Key: tm_privval.FilePVKey{
			Address: priv.PubKey().Address(),
			PubKey:  priv.PubKey(),
			PrivKey: priv,
		},
		LastSignState: tm_privval.FilePVLastSignState{
			Height:    0,
			Round:     0,
			Step:      0,
			Signature: []byte{},
			SignBytes: []byte{},
		},
	}
}

func mockSCFilePV(t *testing.T) *SCFilePV {
	t.Helper()
	return NewSCFilePV(
		log.New(ioutil.Discard, "", 0),
		testConfig(t),
		testState(t),
		testFilePV(t),
		&http.Server{Addr: fmt.Sprintf(":%v", DefaultHTTPPort)},
	)
}
