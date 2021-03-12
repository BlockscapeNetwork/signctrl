package types

import (
	"errors"
	"log"
)

var (
	// ErrThresholdExceeded is returned when the threshold of too many missed blocks in
	// a row is exceeded.
	ErrThresholdExceeded = errors.New("threshold exceeded due to too many blocks missed in a row")

	// ErrMustShutdown is returned when the current signer (rank 1) beeds to update its
	// ranks and must be shut down because rank 1 cannot be promoted anymore.
	ErrMustShutdown = errors.New("node cannot be promoted anymore, so it must be shut down")

	// ErrCounterLocked is returned when the counter for missed blocks in a row is
	// still locked due to SignCTRL not having seen a signed block from rank 1.
	ErrCounterLocked = errors.New("waiting for first commitsig from validator to unlock counter for missed blocks in a row")
)

// SignCtrled defines the functionality of a SignCTRL PrivValidator that monitors the
// blockchain for missed blocks in a row and keeps its rank up to date.
type SignCtrled interface {
	Missed() error
	Reset()
	promote() error
}

// BaseSignCtrled is a base implementation of SignCtrled.
type BaseSignCtrled struct {
	Logger        *log.Logger
	counterLocked bool
	currentHeight int64
	missedInARow  int
	threshold     int
	rank          int

	impl SignCtrled
}

// NewBaseSignCtrled creates a new instance of BaseSignCtrled.
func NewBaseSignCtrled(logger *log.Logger, threshold int, rank int, impl SignCtrled) *BaseSignCtrled {
	return &BaseSignCtrled{
		Logger:        logger,
		counterLocked: true,
		currentHeight: 1,
		threshold:     threshold,
		rank:          rank,
		impl:          impl,
	}
}

// UnlockCounter unlocks the counter for missed blocks in a row.
// This lock is crucial for mitigating the risk of double-signing on startup of the
// validators in the set if they are started up in incorrect order.
func (bsc *BaseSignCtrled) UnlockCounter() {
	if bsc.counterLocked {
		bsc.Logger.Println("[INFO] signctrl: Found first commitsig from validator since fully synced, start counting missed blocks in a row...")
		bsc.counterLocked = false
	}
}

// GetCurrentHeight returns the validator's current height.
func (bsc *BaseSignCtrled) GetCurrentHeight() int64 {
	return bsc.currentHeight
}

// SetCurrentHeight sets the current height to the given value.
func (bsc *BaseSignCtrled) SetCurrentHeight(height int64) {
	bsc.currentHeight = height
}

// GetThreshold returns the threshold of blocks missed in a row that trigger a rank
// update.
func (bsc *BaseSignCtrled) GetThreshold() int {
	return bsc.threshold
}

// GetMissedInARow returns the number of blocks missed in a row.
func (bsc *BaseSignCtrled) GetMissedInARow() int {
	return bsc.missedInARow
}

// GetRank returns the validators current rank.
func (bsc *BaseSignCtrled) GetRank() int {
	return bsc.rank
}

// SetRank sets the validator's rank to the given rank.
func (bsc *BaseSignCtrled) SetRank(rank int) {
	bsc.rank = rank
}

// Missed updates the counter for missed blocks in a row. Errors are returned if...
//
// 1) the threshold of too many blocks missed in a row is exceeded
// 2) the validator's promotion fails
// 3) the counter for missed blocks in a row is still locked
//
// Implements the SignCtrled interface.
func (bsc *BaseSignCtrled) Missed() error {
	if bsc.counterLocked {
		return ErrCounterLocked
	}

	bsc.missedInARow++
	if bsc.missedInARow < bsc.threshold {
		bsc.Logger.Printf("[INFO] signctrl: Missed a block (%v/%v)", bsc.missedInARow, bsc.threshold)
	} else if bsc.missedInARow == bsc.threshold {
		bsc.Logger.Printf("[INFO] signctrl: Missed too many blocks in a row (%v/%v)", bsc.missedInARow, bsc.threshold)
		if err := bsc.promote(); err != nil {
			return err
		}

		// When a rank update due to ErrThresholdExceeded is triggered, it is expected
		// that the next block will not contain the validator's signature. This is due
		// to a block containing the commit of the previous height which we know wasn't
		// signed. Therefore, skip ahead.
		bsc.currentHeight++
		return ErrThresholdExceeded
	}

	return nil
}

// Reset resets the counter for missed blocks in a row to 0.
// Implements the SignCtrled interface.
func (bsc *BaseSignCtrled) Reset() {
	if bsc.missedInARow > 0 {
		bsc.Logger.Println("[DEBUG] signctrl: Reset counter for missed blocks in a row")
		bsc.missedInARow = 0
	}
}

// promote moves the validator up one rank. An error is returned if the validator
// cannot be promoted anymore and it has to be shut down consequently.
// This method is only supposed to be called from within the Missed method and never
// on its own.
// Implements the SignCtrled interface.
func (bsc *BaseSignCtrled) promote() error {
	if bsc.rank == 1 {
		return ErrMustShutdown
	}

	bsc.Logger.Printf("[INFO] signctrl: Promote validator (%v -> %v)", bsc.rank, bsc.rank-1)
	bsc.rank--
	bsc.Reset()

	return nil
}
