package connection

import (
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"syscall"
	"time"

	tm_ed25519 "github.com/tendermint/tendermint/crypto/ed25519"
	tm_p2pconn "github.com/tendermint/tendermint/p2p/conn"
)

var (
	// ErrAbortDial is returned if either SIGINT or SIGTERM are fired into the quit
	// channel.
	ErrAbortDial = errors.New("dialing aborted")

	// RetryDialInterval is the interval in which SignCTRL tries to repeatedly dial
	// the validator. Initially set to 0, then set by retryDialX.
	// Make SignCTRL dial immediately the first time.
	RetryDialInterval = time.Duration(0)
)

// retryDialTCP keeps dialing the given TCP socket address until success, using the
// given connkey for encryption and returns the secret connection.
func retryDialTCP(address string, connkey tm_ed25519.PrivKey, sigs chan os.Signal, logger *log.Logger) (net.Conn, error) {
	for {
		select {
		case <-sigs:
			return nil, ErrAbortDial

		case <-time.After(RetryDialInterval):
			if conn, err := net.Dial("tcp", strings.TrimPrefix(address, "tcp://")); err == nil {
				logger.Println("[INFO] signctrl: Successfully dialed the validator ✓")
				return tm_p2pconn.MakeSecretConnection(conn, connkey)
			}

			// After the first dial, dial in intervals of 1 second.
			RetryDialInterval = time.Second
			logger.Println("[DEBUG] signctrl: Retry dialing...")
		}
	}
}

// retryDialUnix keeps dialing the given unix domain socket address until success and
// returns the connection.
func retryDialUnix(address string, sigs chan os.Signal, logger *log.Logger) (net.Conn, error) {
	addrWithoutProtocol := strings.TrimPrefix(address, "unix://")
	os.RemoveAll(addrWithoutProtocol)

	for {
		select {
		case <-sigs:
			return nil, ErrAbortDial

		case <-time.After(RetryDialInterval):
			unixAddr := &net.UnixAddr{Name: addrWithoutProtocol, Net: "unix"}
			if conn, err := net.DialUnix("unix", nil, unixAddr); err == nil {
				logger.Println("[INFO] signctrl: Successfully dialed the validator ✓")
				return conn, nil
			}

			// After the first dial, dial in intervals of 1 second.
			RetryDialInterval = time.Second
			logger.Println("[DEBUG] signctrl: Retry dialing...")
		}
	}
}

// RetryDial keeps dialing the given address until success and returns the connection.
func RetryDial(cfgDir, address string, logger *log.Logger) (net.Conn, error) {
	logger.Printf("[INFO] signctrl: Dialing %v... (Use Ctrl+C to abort)", address)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	protocol := regexp.MustCompile(`tcp|unix`).FindString(address)
	switch protocol {
	case "tcp":
		// Load the connection key from the config directory which is needed to establish
		// a secret/encrypted connection to the validator.
		connKey, err := LoadConnKey(cfgDir)
		if err != nil {
			return nil, fmt.Errorf("[ERR] signctrl: couldn't load conn.key: %v", err)
		}
		return retryDialTCP(address, connKey, sigs, logger)

	case "unix":
		return retryDialUnix(address, sigs, logger)

	default:
		return nil, fmt.Errorf("unknown protocol in address: %v", protocol)
	}
}
