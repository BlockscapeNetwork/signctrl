package connection

import (
	"crypto/ed25519"
	"crypto/rand"
	"fmt"
	"log"
	"net"
	"os"
	"testing"

	tmcrypto "github.com/tendermint/tendermint/crypto/ed25519"
	p2pconn "github.com/tendermint/tendermint/p2p/conn"
)

func TestRetrySecretDial(t *testing.T) {
	// Generate private key used to make a secret connection.
	_, priv, _ := ed25519.GenerateKey(rand.Reader)

	// Error channel used to report errors from the goroutine
	// back to the main test routine.
	errCh := make(chan error)

	// Start tcp socket listener.
	go func() {
		l, err := net.Listen("tcp", "127.0.0.1:3000")
		if err != nil {
			errCh <- fmt.Errorf("Expected err to be nil, instead got: %v", err)
			return
		}
		defer l.Close()

		conn, err := l.Accept()
		if err != nil {
			errCh <- fmt.Errorf("Expected err to be nil, instead got: %v", err)
			return
		}

		secretConn, err := p2pconn.MakeSecretConnection(conn, tmcrypto.PrivKey(priv))
		if err != nil {
			errCh <- fmt.Errorf("Expected err to be nil, instead got: %v", err)
			return
		}
		defer secretConn.Close()
	}()

	rwc := NewReadWriteConn()
	var err error

	rwc.SecretConn, err = RetrySecretDial("tcp", "127.0.0.1:3000", priv, log.New(os.Stderr, "", 0))
	if err != nil {
		t.Fatalf("Expected err to be nil, instead got: %v", err)
	}
	defer rwc.SecretConn.Close()
}
