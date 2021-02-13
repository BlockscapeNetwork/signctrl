package privval

import (
	"crypto/ed25519"
	"crypto/rand"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/BlockscapeNetwork/signctrl/config"
	"github.com/BlockscapeNetwork/signctrl/connection"
	tmcrypto "github.com/tendermint/tendermint/crypto"
	tmed25519 "github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/crypto/tmhash"
	tmjson "github.com/tendermint/tendermint/libs/json"
	tmlog "github.com/tendermint/tendermint/libs/log"
	"github.com/tendermint/tendermint/libs/protoio"
	tmprivval "github.com/tendermint/tendermint/privval"
	tmprototypes "github.com/tendermint/tendermint/proto/tendermint/types"
	coretypes "github.com/tendermint/tendermint/rpc/core/types"
	tmtypes "github.com/tendermint/tendermint/types"
)

func testVote() *tmprototypes.Vote {
	return &tmprototypes.Vote{
		Type:   tmprototypes.PrecommitType,
		Height: 11,
		Round:  2,
		BlockID: tmprototypes.BlockID{
			Hash: tmhash.Sum([]byte("BlockIDHash")),
			PartSetHeader: tmprototypes.PartSetHeader{
				Total: 65536,
				Hash:  tmhash.Sum([]byte("BlockIDPartSetHeaderHash")),
			},
		},
		Timestamp:        time.Now(),
		ValidatorAddress: tmcrypto.AddressHash([]byte("ValidatorAddress")),
		ValidatorIndex:   1337,
		Signature:        []byte{},
	}
}

func testProposal() *tmprototypes.Proposal {
	return &tmprototypes.Proposal{
		Type:     tmprototypes.ProposalType,
		Height:   23,
		Round:    1,
		PolRound: 2,
		BlockID: tmprototypes.BlockID{
			Hash: tmhash.Sum([]byte("BlockIDHash")),
			PartSetHeader: tmprototypes.PartSetHeader{
				Total: 1,
				Hash:  tmhash.Sum([]byte("BlockIDPartSetHeaderHash")),
			},
		},
		Timestamp: time.Now(),
		Signature: []byte("Signature"),
	}
}

func testCommit() *connection.CommitRPCResponse {
	return &connection.CommitRPCResponse{
		Result: &coretypes.ResultCommit{
			SignedHeader: tmtypes.SignedHeader{
				Commit: &tmtypes.Commit{
					Signatures: []tmtypes.CommitSig{
						{
							ValidatorAddress: []byte("VAL-1-ADDR"),
							Signature:        []byte("VAL-1-SIG"),
						},
						{
							ValidatorAddress: []byte("VAL-2-ADDR"),
							Signature:        []byte("VAL-2-SIG"),
						},
					},
				},
			},
		},
	}
}

func testPrivValidatorKey() string {
	return `{
"address": "5BCD69E0178E0E6C6F96F541B265CAE3178611AE",
"pub_key": {
  "type": "tendermint/PubKeyEd25519",
  "value": "KwddNyi18Ta7tPs6xwfM79O3waMn1+aJuB6GyGQjYuY="
},
"priv_key": {
  "type": "tendermint/PrivKeyEd25519",
  "value": "XQpf+QIrfT/3v0yLquLhfJ5dUaQfJ+ScLYoPzjpUuTkrB103KLXxNru0+zrHB8zv07fBoyfX5om4HobIZCNi5g=="
  }
}`
}

func testPrivValidatorState() string {
	return `{
  "height": "0",
  "round": 0,
  "step": 0
}`
}

func testValidConfig() *config.Config {
	return &config.Config{
		Init: config.InitConfig{
			LogLevel:               "INFO",
			SetSize:                2,
			Threshold:              10,
			Rank:                   1,
			ValidatorListenAddr:    "127.0.0.1:4000",
			ValidatorListenAddrRPC: "127.0.0.1:26657",
		},
		FilePV: config.FilePVConfig{
			ChainID:       "testchain",
			KeyFilePath:   "./priv_validator_key.json",
			StateFilePath: "./priv_validator_state.json",
		},
	}
}

func TestMissed(t *testing.T) {
	pm := NewSCFilePV()
	pm.Logger = log.New(os.Stderr, "", 0)
	pm.Config.Init.Threshold = 3
	pm.CounterUnlocked = true

	for i := 0; i < pm.Config.Init.Threshold-1; i++ {
		if err := pm.Missed(); err != nil {
			t.Errorf("Expected err to be nil, instead got: %v", err)
		}
	}

	if err := pm.Missed(); err == nil {
		t.Error("Expected err, instead got nil")
	}
}

func TestReset(t *testing.T) {
	pm := NewSCFilePV()
	pm.Logger = log.New(os.Stderr, "", 0)
	pm.MissedInARow = 3
	pm.Reset()

	if pm.MissedInARow != 0 {
		t.Errorf("Expected MissedInARow to be 0, instead got: %v", pm.MissedInARow)
	}
}

func TestUpdate(t *testing.T) {
	pm := NewSCFilePV()
	pm.Logger = log.New(os.Stderr, "", 0)
	pm.Config.Init.Rank = 2
	pm.Update()

	if pm.Config.Init.Rank != 1 {
		t.Errorf("Expected rank to be 1, instead got: %v", pm.Config.Init.Rank)
	}
}

func TestGetPubKey(t *testing.T) {
	pm := NewSCFilePV()
	priv := tmed25519.GenPrivKey()
	pm.FilePV = tmprivval.NewFilePV(priv, "", "")

	// There is currently no way to make GetPubKey return an error.
	if _, err := pm.GetPubKey(); err != nil {
		t.Errorf("Expected err to be nil, instead got: %v", err)
	}
}

func TestSignVote(t *testing.T) {
	pm := NewSCFilePV()
	priv := tmed25519.GenPrivKey()
	pm.FilePV = tmprivval.NewFilePV(priv, "./priv_validator_key.json", "./priv_validator_state.json")
	defer os.Remove("./priv_validator_key.json")
	defer os.Remove("./priv_validator_state.json")

	if err := pm.SignVote("testchain", testVote()); err != nil {
		t.Errorf("Expected err to be nil, instead got: %v", err)
	}

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("SignVote should have panicked")
		}
	}()

	pm.SignVote("testchain", nil) // panic
}

func TestSignProposal(t *testing.T) {
	pm := NewSCFilePV()
	priv := tmed25519.GenPrivKey()
	pm.FilePV = tmprivval.NewFilePV(priv, "./priv_validator_key.json", "./priv_validator_state.json")
	defer os.Remove("./priv_validator_key.json")
	defer os.Remove("./priv_validator_state.json")

	if err := pm.SignProposal("testchain", testProposal()); err != nil {
		t.Errorf("Expected err to be nil, instead got: %v", err)
	}

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("SignVote should have panicked")
		}
	}()

	pm.SignProposal("testchain", nil) // panic
}

func testSignerClient(listenAddr, chainID string, quit chan<- error) {
	pve, err := tmprivval.NewSignerListener(listenAddr, tmlog.NewNopLogger())
	if err != nil {
		quit <- err
		return
	}

	sc, err := tmprivval.NewSignerClient(pve, chainID)
	if err != nil {
		quit <- err
		return
	}
	defer sc.Close()

	_, err = sc.GetPubKey()
	if err != nil {
		quit <- err
		return
	}

	if err := sc.SignVote(chainID, testVote()); err != nil {
		quit <- err
		return
	}

	if err := sc.SignProposal(chainID, testProposal()); err != nil {
		quit <- err
		return
	}

	quit <- nil
	return
}

func testSignCtrled(protocol, dialAddr string, privKey ed25519.PrivateKey, quit chan<- error, sigCh chan os.Signal) {
	ioutil.WriteFile("./priv_validator_key.json", []byte(testPrivValidatorKey()), 0644)
	ioutil.WriteFile("./priv_validator_state.json", []byte(testPrivValidatorState()), 0644)
	defer os.Remove("./priv_validator_key.json")
	defer os.Remove("./priv_validator_state.json")

	pm := NewSCFilePV()
	pm.Logger = log.New(os.Stderr, "", 0)
	pm.Config = testValidConfig()
	pm.FilePV = tmprivval.LoadOrGenFilePV(pm.Config.FilePV.KeyFilePath, pm.Config.FilePV.StateFilePath)

	rwc := connection.NewReadWriteConn()

	var err error
	rwc.SecretConn, err = connection.RetrySecretDial("tcp", "127.0.0.1:3000", privKey, pm.Logger)
	if err != nil {
		quit <- err
		return
	}

	rwc.Reader = protoio.NewDelimitedReader(rwc.SecretConn, 64<<10)
	rwc.Writer = protoio.NewDelimitedWriter(rwc.SecretConn)

	pubkey, err := pm.GetPubKey()
	if err != nil {
		quit <- err
		return
	}

	pm.Run(rwc, pubkey, sigCh)
}

func testCommitEndpoint(quit chan<- error) {
	http.HandleFunc("/commit", func(w http.ResponseWriter, r *http.Request) {
		height, ok := r.URL.Query()["height"]
		if !ok || len(height[0]) < 1 {
			quit <- errors.New("URL param 'height' is missing")
			return
		}

		bytes, err := tmjson.Marshal(testCommit())
		if err != nil {
			w.Write(nil)
			return
		}

		w.Write(bytes)
	})

	if err := http.ListenAndServe(":26657", nil); err != nil {
		quit <- err
		return
	}
}

func TestRun(t *testing.T) {
	quitCh := make(chan error)
	sigCh := make(chan os.Signal)
	_, priv, _ := ed25519.GenerateKey(rand.Reader)

	go testSignerClient("127.0.0.1:3000", "testchain", quitCh)
	go testCommitEndpoint(quitCh)
	go testSignCtrled("tcp", "127.0.0.1:3000", priv, quitCh, sigCh)

	if err := <-quitCh; err != nil {
		t.Fatalf("Expected err to be nil, instead got: %v", err)
	}

	sigCh <- syscall.SIGINT
}
