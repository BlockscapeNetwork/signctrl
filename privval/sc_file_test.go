package privval

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"testing"

	"github.com/BlockscapeNetwork/signctrl/config"
	"github.com/stretchr/testify/assert"
	tm_crypto "github.com/tendermint/tendermint/crypto"
	tm_ed25519 "github.com/tendermint/tendermint/crypto/ed25519"
	tm_privval "github.com/tendermint/tendermint/privval"
	tm_prototypes "github.com/tendermint/tendermint/proto/tendermint/types"
	tm_types "github.com/tendermint/tendermint/types"
)

type TestFilePV struct{}

func NewTestFilePV() *TestFilePV {
	return &TestFilePV{}
}

func (tpv *TestFilePV) GetPubKey() (tm_crypto.PubKey, error) {
	priv := tm_ed25519.GenPrivKey()
	return priv.PubKey(), nil
}

func (tpv *TestFilePV) SignVote(chainID string, vote *tm_prototypes.Vote) error {
	return errors.New("")
}

func (tpv *TestFilePV) SignProposal(chainID string, proposal *tm_prototypes.Proposal) error {
	return errors.New("")
}

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

func testState(t *testing.T) config.State {
	t.Helper()
	return config.State{
		LastHeight: 1,
		LastRank:   1,
	}
}

func testFilePV(t *testing.T) tm_types.PrivValidator {
	t.Helper()
	priv := tm_ed25519.GenPrivKey()
	return &tm_privval.FilePV{
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

func TestKeyFilePath(t *testing.T) {
	path := KeyFilePath("/tmp")
	assert.Equal(t, "/tmp/priv_validator_key.json", path)
}

func TestStateFilePath(t *testing.T) {
	path := StateFilePath("/tmp")
	assert.Equal(t, "/tmp/priv_validator_state.json", path)
}
