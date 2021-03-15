package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type testSignCtrled struct {
	BaseSignCtrled
}

func TestMissed(t *testing.T) {
	sc := &testSignCtrled{}
	sc.BaseSignCtrled = *NewBaseSignCtrled(nil, 2, 1, sc)

	sc.UnlockCounter()
	err := sc.Missed()
	assert.NoError(t, err)
	assert.Equal(t, 1, sc.GetMissedInARow())
	assert.Equal(t, 1, sc.GetRank())

	sc.LockCounter()
	err = sc.Missed()
	assert.ErrorIs(t, ErrCounterLocked, err)
	assert.Equal(t, 1, sc.GetMissedInARow())
	assert.Equal(t, 1, sc.GetRank())
}

func TestThresholdExceeded(t *testing.T) {
	sc := &testSignCtrled{}
	sc.BaseSignCtrled = *NewBaseSignCtrled(nil, 1, 2, sc)

	sc.UnlockCounter()
	err := sc.Missed()
	assert.ErrorIs(t, ErrThresholdExceeded, err)
	assert.Equal(t, 0, sc.GetMissedInARow())
	assert.Equal(t, 1, sc.GetRank())
}

func TestReset(t *testing.T) {
	sc := &testSignCtrled{}
	sc.BaseSignCtrled = *NewBaseSignCtrled(nil, 2, 1, sc)
	sc.missedInARow = 1

	sc.UnlockCounter()
	sc.Reset()
	assert.Equal(t, 0, sc.GetMissedInARow())
}

func TestPromote(t *testing.T) {
	sc := &testSignCtrled{}
	sc.BaseSignCtrled = *NewBaseSignCtrled(nil, 1, 1, sc)

	sc.UnlockCounter()
	err := sc.Missed()
	assert.ErrorIs(t, ErrMustShutdown, err)
}
