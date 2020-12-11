package types

import "github.com/tendermint/tendermint/proto/tendermint/types"

// State tracks the height, round and step of the last message
// received from Tendermint.
type State struct {
	Height int64
	Round  int32
	Step   types.SignedMsgType
}

// TODO
