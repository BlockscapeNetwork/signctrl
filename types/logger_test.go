package types

import (
	"os"
	"testing"
)

func TestSyncLoggerDebug(t *testing.T) {
	sl := NewSyncLogger(os.Stderr, "", 0)
	sl.Debug("Debug test msg")
	// Output:
	// [DEBUG] signctrl: Debug test msg
}

func TestSyncLoggerInfo(t *testing.T) {
	sl := NewSyncLogger(os.Stderr, "", 0)
	sl.Info("Info test msg")
	// Output:
	// [INFO] signctrl: Info test msg
}

func TestSyncLoggerWarn(t *testing.T) {
	sl := NewSyncLogger(os.Stderr, "", 0)
	sl.Warn("Warn test msg")
	// Output:
	// [WARN] signctrl: Debug test msg
}

func TestSyncLoggerError(t *testing.T) {
	sl := NewSyncLogger(os.Stderr, "", 0)
	sl.Error("Error test msg")
	// Output:
	// [ERR] signctrl: Debug test msg
}
