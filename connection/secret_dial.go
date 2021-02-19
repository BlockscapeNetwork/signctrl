package connection

import (
	"errors"
	"log"
	"net"
	"os"
	"strings"
	"time"

	tm_ed25519 "github.com/tendermint/tendermint/crypto/ed25519"
	tm_p2pconn "github.com/tendermint/tendermint/p2p/conn"
)

const (
	// RetryDialInterval is the interval in which SignCTRL tries to repeatedly dial
	// the validator.
	RetryDialInterval = 500 * time.Millisecond
)

var (
	// ErrAbortDial is returned if either SIGINT or SIGTERM are fired into the quit
	// channel.
	ErrAbortDial = errors.New("aborted dialing the validator")
)

// RetrySecretDialTCP keeps dialing the given TCP address until success, using the
// given privkey for encryption and returns the secret connection.
func RetrySecretDialTCP(address string, privkey tm_ed25519.PrivKey, logger *log.Logger) (net.Conn, error) {
	logger.Printf("[INFO] signctrl: Dialing %v... (Press Ctrl+C to abort)", address)
	quit := make(chan os.Signal, 1)
	defer close(quit)

	for {
		select {
		case <-quit:
			return nil, ErrAbortDial

		default:
			if conn, err := net.Dial("tcp", strings.TrimPrefix(address, "tcp://")); err == nil {
				logger.Println("[INFO] signctrl: Successfully dialed the validator")
				return tm_p2pconn.MakeSecretConnection(conn, privkey)
			}
			time.Sleep(RetryDialInterval)
		}
	}
}
