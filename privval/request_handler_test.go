package privval

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/BlockscapeNetwork/signctrl/rpc"
	"github.com/BlockscapeNetwork/signctrl/types"
	"github.com/stretchr/testify/assert"
	tm_crypto "github.com/tendermint/tendermint/crypto"
	tm_hash "github.com/tendermint/tendermint/crypto/tmhash"
	tm_json "github.com/tendermint/tendermint/libs/json"
	tm_privval "github.com/tendermint/tendermint/privval"
	tm_privvalproto "github.com/tendermint/tendermint/proto/tendermint/privval"
	tm_prototypes "github.com/tendermint/tendermint/proto/tendermint/types"
	tm_coretypes "github.com/tendermint/tendermint/rpc/core/types"
	tm_types "github.com/tendermint/tendermint/types"
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

func TestWrapMsg(t *testing.T) {
	msg := wrapMsg(&tm_privvalproto.Message{})
	assert.NotNil(t, msg)

	msg = wrapMsg(&tm_privvalproto.PingRequest{})
	assert.NotNil(t, msg)

	msg = wrapMsg(&tm_privvalproto.PingResponse{})
	assert.NotNil(t, msg)

	msg = wrapMsg(&tm_privvalproto.PubKeyRequest{})
	assert.NotNil(t, msg)

	msg = wrapMsg(&tm_privvalproto.PubKeyResponse{})
	assert.NotNil(t, msg)

	msg = wrapMsg(&tm_privvalproto.SignVoteRequest{})
	assert.NotNil(t, msg)

	msg = wrapMsg(&tm_privvalproto.SignedVoteResponse{})
	assert.NotNil(t, msg)

	msg = wrapMsg(&tm_privvalproto.SignProposalRequest{})
	assert.NotNil(t, msg)

	msg = wrapMsg(&tm_privvalproto.SignedProposalResponse{})
	assert.NotNil(t, msg)

	assert.Panics(t, func() { wrapMsg(nil) })
}

func testCommitSigs(t *testing.T) *[]tm_types.CommitSig {
	t.Helper()
	return &[]tm_types.CommitSig{
		{
			ValidatorAddress: []byte("ALPHA-ADDR"),
			Signature:        []byte("ALPHA-SIG"),
		},
		{
			ValidatorAddress: []byte("BETA-ADDR"),
			Signature:        []byte("BETA-SIG"),
		},
	}
}

func TestHasSignedCommit(t *testing.T) {
	signed := hasSignedCommit([]byte("ALPHA-ADDR"), testCommitSigs(t))
	assert.True(t, signed)

	signed = hasSignedCommit([]byte("BETA-SIG"), testCommitSigs(t))
	assert.False(t, signed)

	signed = hasSignedCommit([]byte("GAMMA"), testCommitSigs(t))
	assert.False(t, signed)
}

func TestIsRankUpToDate(t *testing.T) {
	upToDate := isRankUpToDate(2, 1, 1)
	assert.True(t, upToDate)

	upToDate = isRankUpToDate(3, 1, 1)
	assert.False(t, upToDate)
}

func testPingRequest(t *testing.T) *tm_privvalproto.Message {
	t.Helper()
	return &tm_privvalproto.Message{
		Sum: &tm_privvalproto.Message_PingRequest{},
	}
}

func TestHandlePingRequest(t *testing.T) {
	pv := mockSCFilePV(t)
	msg, err := HandleRequest(context.Background(), testPingRequest(t), pv)
	assert.NotNil(t, msg)
	assert.NoError(t, err)
}

func testPubKeyRequest(t *testing.T) *tm_privvalproto.Message {
	t.Helper()
	return &tm_privvalproto.Message{
		Sum: &tm_privvalproto.Message_PubKeyRequest{
			PubKeyRequest: &tm_privvalproto.PubKeyRequest{
				ChainId: "testchain",
			},
		},
	}
}

func TestHandlePubKeyRequest_WrongChainID(t *testing.T) {
	// Initialize mock SCFilePV with valid values.
	pv := mockSCFilePV(t)

	// Wrong chain ID.
	pv.Config.Privval.ChainID = "wrongchain"

	// Handle request.
	msg, err := HandleRequest(context.Background(), testPubKeyRequest(t), pv)
	assert.NotNil(t, msg)
	assert.Error(t, err)
}

func TestHandlePubKeyRequest_InvalidPubKey(t *testing.T) {
	// Initialize mock SCFilePV with valid values.
	pv := mockSCFilePV(t)

	// Nil pubkey.
	tmpv, ok := pv.TMFilePV.(*tm_privval.FilePV)
	assert.True(t, ok)
	tmpv.Key.PubKey = nil

	// Handle request.
	msg, err := HandleRequest(context.Background(), testPubKeyRequest(t), pv)
	assert.NotNil(t, msg)
	assert.Error(t, err)
}

func TestHandlePubKeyRequest_Valid(t *testing.T) {
	// Initialize mock SCFilePV with valid values.
	pv := mockSCFilePV(t)

	// Handle request.
	msg, err := HandleRequest(context.Background(), testPubKeyRequest(t), pv)
	assert.NotNil(t, msg)
	assert.NoError(t, err)
}

func testVote(t *testing.T) *tm_prototypes.Vote {
	t.Helper()
	return &tm_prototypes.Vote{
		Type:   tm_prototypes.PrecommitType,
		Height: 2,
		Round:  1,
		BlockID: tm_prototypes.BlockID{
			Hash: tm_hash.Sum([]byte("BlockIDHash")),
			PartSetHeader: tm_prototypes.PartSetHeader{
				Total: 65536,
				Hash:  tm_hash.Sum([]byte("BlockIDPartSetHeaderHash")),
			},
		},
		Timestamp:        time.Now(),
		ValidatorAddress: tm_crypto.AddressHash([]byte("ValidatorAddress")),
		ValidatorIndex:   1337,
		Signature:        []byte{},
	}
}

func testSignVoteRequest(t *testing.T) *tm_privvalproto.Message {
	t.Helper()
	return &tm_privvalproto.Message{
		Sum: &tm_privvalproto.Message_SignVoteRequest{
			SignVoteRequest: &tm_privvalproto.SignVoteRequest{
				Vote:    testVote(t),
				ChainId: "testchain",
			},
		},
	}
}

func testProposal(t *testing.T) *tm_prototypes.Proposal {
	t.Helper()
	return &tm_prototypes.Proposal{
		Type:     tm_prototypes.ProposalType,
		Height:   2,
		Round:    1,
		PolRound: 2,
		BlockID: tm_prototypes.BlockID{
			Hash: tm_hash.Sum([]byte("BlockIDHash")),
			PartSetHeader: tm_prototypes.PartSetHeader{
				Total: 1,
				Hash:  tm_hash.Sum([]byte("BlockIDPartSetHeaderHash")),
			},
		},
		Timestamp: time.Now(),
		Signature: []byte("Signature"),
	}
}

func testSignProposalRequest(t *testing.T) *tm_privvalproto.Message {
	t.Helper()
	return &tm_privvalproto.Message{
		Sum: &tm_privvalproto.Message_SignProposalRequest{
			SignProposalRequest: &tm_privvalproto.SignProposalRequest{
				Proposal: testProposal(t),
				ChainId:  "testchain",
			},
		},
	}
}

func TestGetSharedSignRequestData(t *testing.T) {
	data := getSharedSignRequestData(testSignVoteRequest(t))
	assert.NotNil(t, data)
	assert.Equal(t, testSignVoteRequest(t).GetSignVoteRequest().ChainId, data.chainID)
	assert.Equal(t, testSignVoteRequest(t).GetSignVoteRequest().Vote.Type, data.msgType)
	assert.Equal(t, testSignVoteRequest(t).GetSignVoteRequest().Vote.Height, data.height)

	data = getSharedSignRequestData(testSignProposalRequest(t))
	assert.NotNil(t, data)
	assert.Equal(t, testSignProposalRequest(t).GetSignProposalRequest().ChainId, data.chainID)
	assert.Equal(t, testSignProposalRequest(t).GetSignProposalRequest().Proposal.Type, data.msgType)
	assert.Equal(t, testSignProposalRequest(t).GetSignProposalRequest().Proposal.Height, data.height)
}

func TestBuildResponse(t *testing.T) {
	resp := buildResponse(wrapMsg(testSignVoteRequest(t)), nil)
	assert.NotNil(t, resp)
	assert.IsType(t, &tm_privvalproto.Message_SignedVoteResponse{}, resp.GetSum())

	resp = buildResponse(wrapMsg(testSignProposalRequest(t)), nil)
	assert.NotNil(t, resp)
	assert.IsType(t, &tm_privvalproto.Message_SignedProposalResponse{}, resp.GetSum())

	resp = buildResponse(wrapMsg(&tm_privvalproto.Message{}), nil)
	assert.Nil(t, resp)
}

func testBlockResult(t *testing.T) *rpc.BlockResult {
	t.Helper()
	return &rpc.BlockResult{
		Result: &tm_coretypes.ResultBlock{
			Block: &tm_types.Block{
				LastCommit: &tm_types.Commit{
					Signatures: []tm_types.CommitSig{
						{
							ValidatorAddress: []byte("ALPHA-ADDR"),
							Signature:        []byte("ALPHA-SIG"),
						},
						{
							ValidatorAddress: []byte("BETA-ADDR"),
							Signature:        []byte("BETA-SIG"),
						},
					},
				},
			},
		},
	}
}

func testBlockEndpoint(t *testing.T, port int, result *rpc.BlockResult, quitCh chan struct{}) {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/block", func(rw http.ResponseWriter, r *http.Request) {
		bytes, _ := tm_json.Marshal(result)
		_, _ = rw.Write(bytes)
	})

	server := http.Server{Addr: fmt.Sprintf(":%v", port), Handler: mux}
	defer server.Close()

	listener, _ := net.Listen("tcp", server.Addr)
	go func() {
		_ = server.Serve(listener)
	}()
	<-quitCh
}

func TestHandleSignRequest(t *testing.T) {
	// Initialize mock SCFilePV with valid values.
	pv := mockSCFilePV(t)

	// Start mock endpoint for the block query.
	port, _ := getFreePort(t)
	pv.Config.Base.ValidatorListenAddressRPC = fmt.Sprintf("tcp://127.0.0.1:%v", port)
	quitCh := make(chan struct{})
	go testBlockEndpoint(t, port, testBlockResult(t), quitCh)
	defer close(quitCh)

	// Initialize new file signer.
	tmpv, ok := pv.TMFilePV.(*tm_privval.FilePV)
	assert.True(t, ok)

	pv.TMFilePV = tm_privval.NewFilePV(tmpv.Key.PrivKey, "./priv_validator_key.json", "./priv_validator_state.json")
	defer os.Remove("./priv_validator_key.json")
	defer os.Remove("./priv_validator_state.json")

	// Handle the request.
	msg, err := HandleRequest(context.Background(), testSignVoteRequest(t), pv)
	assert.NotNil(t, msg)
	assert.NoError(t, err)
}

func TestHandleSignRequest_WrongChainID(t *testing.T) {
	// Initialize mock SCFilePV with valid values.
	pv := mockSCFilePV(t)

	// The testSignVoteRequest has "testchain", so change it that they don't match.
	pv.Config.Privval.ChainID = "wrongchain"

	// Handle the request.
	msg, err := HandleRequest(context.Background(), testSignVoteRequest(t), pv)
	assert.NotNil(t, msg)
	assert.Error(t, err)
}

func TestHandleSignRequest_ObsoleteRank(t *testing.T) {
	// Initialize mock SCFilePV with valid values.
	pv := mockSCFilePV(t)

	// Make rank obsolete by making the gap between the requested vote height
	// and the threshold {threshold+2} wide.
	req := testSignVoteRequest(t)
	req.GetSignVoteRequest().Vote.Height = int64(pv.GetThreshold()) + 2

	// Handle the request.
	msg, err := HandleRequest(context.Background(), req, pv)
	assert.NotNil(t, msg)
	assert.Error(t, err)
}

func TestHandleSignRequest_QueryBlockErr(t *testing.T) {
	// Initialize mock SCFilePV with valid values.
	pv := mockSCFilePV(t)

	// There's no HTTP server running to process the block request, so
	// QueryBlock is going to return an error.

	// Handle the request.
	msg, err := HandleRequest(context.Background(), testSignVoteRequest(t), pv)
	assert.NotNil(t, msg)
	assert.Error(t, err)
}

func TestHandleSignRequest_Missed(t *testing.T) {
	// Initialize mock SCFilePV with valid values.
	pv := mockSCFilePV(t)
	pv.UnlockCounter()

	// Start mock endpoint for the block query.
	port, _ := getFreePort(t)
	pv.Config.Base.ValidatorListenAddressRPC = fmt.Sprintf("tcp://127.0.0.1:%v", port)
	quitCh := make(chan struct{})
	go testBlockEndpoint(t, port, testBlockResult(t), quitCh)
	defer close(quitCh)

	// Initialize new file signer.
	tmpv, ok := pv.TMFilePV.(*tm_privval.FilePV)
	assert.True(t, ok)

	pv.TMFilePV = tm_privval.NewFilePV(tmpv.Key.PrivKey, "./priv_validator_key.json", "./priv_validator_state.json")
	defer os.Remove("./priv_validator_key.json")
	defer os.Remove("./priv_validator_state.json")

	// The testBlockResult doesn't contain the validator's commitsig, so it will be
	// marked as missed. The threshold will not be exceeded, though, so the vote will
	// be signed.

	// Handle the request.
	msg, err := HandleRequest(context.Background(), testSignVoteRequest(t), pv)
	assert.NotNil(t, msg)
	assert.NoError(t, err)
}

func TestHandleSignRequest_MustShutdown(t *testing.T) {
	// Initialize mock SCFilePV with valid values.
	pv := mockSCFilePV(t)

	// Set threshold to 1, so the next missed block exceeds it, and also rank to 1
	// so the promotion fails.
	pv.BaseSignCtrled = *types.NewBaseSignCtrled(
		pv.Logger,
		1, // Threshold
		1, // Rank
		pv,
	)
	pv.UnlockCounter()

	// Start mock endpoint for the block query.
	port, _ := getFreePort(t)
	pv.Config.Base.ValidatorListenAddressRPC = fmt.Sprintf("tcp://127.0.0.1:%v", port)
	quitCh := make(chan struct{})
	go testBlockEndpoint(t, port, testBlockResult(t), quitCh)
	defer close(quitCh)

	// Handle the request.
	msg, err := HandleRequest(context.Background(), testSignVoteRequest(t), pv)
	assert.NotNil(t, msg)
	assert.Error(t, err)
}

func TestHandleSignRequest_RankTooLow(t *testing.T) {
	// Initialize mock SCFilePV with valid values.
	pv := mockSCFilePV(t)
	pv.BaseSignCtrled.SetRank(2)
	pv.UnlockCounter()

	tmpv, ok := pv.TMFilePV.(*tm_privval.FilePV)
	assert.True(t, ok)

	// Add the validator's address to the commitsigs.
	quitCh := make(chan struct{})
	br := testBlockResult(t)
	br.Result.Block.LastCommit.Signatures = []tm_types.CommitSig{
		{
			ValidatorAddress: tmpv.GetAddress(),
		},
	}

	// Start mock endpoint for the block query.
	port, _ := getFreePort(t)
	pv.Config.Base.ValidatorListenAddressRPC = fmt.Sprintf("tcp://127.0.0.1:%v", port)
	go testBlockEndpoint(t, port, br, quitCh)
	defer close(quitCh)

	// Handle the request.
	msg, err := HandleRequest(context.Background(), testSignVoteRequest(t), pv)
	assert.NotNil(t, msg)
	assert.Error(t, err)
	assert.Equal(t, 0, pv.GetMissedInARow())
}

func TestHandleSignRequest_SignVoteErr(t *testing.T) {
	// Initialize mock SCFilePV with valid values.
	pv := mockSCFilePV(t)
	pv.UnlockCounter()

	tmpv, ok := pv.TMFilePV.(*tm_privval.FilePV)
	assert.True(t, ok)

	// Add the validator's address to the commitsigs.
	quitCh := make(chan struct{})
	br := testBlockResult(t)
	br.Result.Block.LastCommit.Signatures = []tm_types.CommitSig{
		{ValidatorAddress: tmpv.GetAddress()},
	}

	// Start mock endpoint for the block query.
	port, _ := getFreePort(t)
	pv.Config.Base.ValidatorListenAddressRPC = fmt.Sprintf("tcp://127.0.0.1:%v", port)
	go testBlockEndpoint(t, port, br, quitCh)
	defer close(quitCh)

	// Initialize new file signer.
	pv.TMFilePV = NewTestFilePV()

	// Handle the request.
	msg, err := HandleRequest(context.Background(), testSignVoteRequest(t), pv)
	assert.NotNil(t, msg)
	assert.Error(t, err)
}

func TestHandleRequest_UnknownMessage(t *testing.T) {
	msg, err := HandleRequest(context.Background(), &tm_privvalproto.Message{}, nil)
	assert.Nil(t, msg)
	assert.Error(t, err)
}
