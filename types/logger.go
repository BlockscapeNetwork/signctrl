package types

import (
	"fmt"
	"io"
	"log"
	"sync"

	"github.com/hashicorp/logutils"
)

var (
	// LogLevels defines the loglevels for SignCTRL logs.
	LogLevels = []logutils.LogLevel{"DEBUG", "INFO", "WARN", "ERR"}
)

// SyncLogger wraps a standard log.Logger and makes it synchronous.
type SyncLogger struct {
	sync.Mutex
	logger *log.Logger
}

// NewSyncLogger creates a new synchronous logger.
func NewSyncLogger(out io.Writer, prefix string, flag int) *SyncLogger {
	return &SyncLogger{logger: log.New(out, prefix, flag)}
}

// SetOutput sets the output destination for the standard logger.
func (sl *SyncLogger) SetOutput(w io.Writer) {
	sl.logger.SetOutput(w)
}

// Debug calls sl.Output to print a debug message to the logger.
func (sl *SyncLogger) Debug(format string, v ...interface{}) {
	sl.Lock()
	defer sl.Unlock()
	taggedFormat := fmt.Sprintf("[DEBUG] signctrl: %v", format)
	_ = sl.logger.Output(2, fmt.Sprintf(taggedFormat, v...))
}

// Info calls sl.Output to print an info message to the logger.
func (sl *SyncLogger) Info(format string, v ...interface{}) {
	sl.Lock()
	defer sl.Unlock()
	taggedFormat := fmt.Sprintf("[INFO]  signctrl: %v", format)
	_ = sl.logger.Output(2, fmt.Sprintf(taggedFormat, v...))
}

// Warn calls sl.Output to print a warning message to the logger.
func (sl *SyncLogger) Warn(format string, v ...interface{}) {
	sl.Lock()
	defer sl.Unlock()
	taggedFormat := fmt.Sprintf("[WARN]  signctrl: %v", format)
	_ = sl.logger.Output(2, fmt.Sprintf(taggedFormat, v...))
}

// Error calls sl.Output to print an error message to the logger.
func (sl *SyncLogger) Error(format string, v ...interface{}) {
	sl.Lock()
	defer sl.Unlock()
	taggedFormat := fmt.Sprintf("[ERR]   signctrl: %v", format)
	_ = sl.logger.Output(2, fmt.Sprintf(taggedFormat, v...))
}
