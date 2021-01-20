package privval

import (
	"bytes"

	"github.com/tendermint/tendermint/types"
)

// hasSignedCommit checks whether the given validator address has a commitsig
// in the provided commitsigs.
func hasSignedCommit(valaddr types.Address, commitsigs *[]types.CommitSig) bool {
	for _, commitsig := range *commitsigs {
		if cmp := bytes.Compare(commitsig.ValidatorAddress, valaddr); cmp == 0 && commitsig.Signature != nil {
			return true
		}
	}

	return false
}
