package connection

import (
	"crypto/ed25519"
	"log"
	"net"
	"time"

	tmcrypto "github.com/tendermint/tendermint/crypto/ed25519"
	"github.com/tendermint/tendermint/libs/protoio"
	p2pconn "github.com/tendermint/tendermint/p2p/conn"
)

// ReadWriteConn holds the secret connection, the reader and the writer.
type ReadWriteConn struct {
	// SecretConn holds the secret connection for communication between
	// SignCTRL and Tendermint.
	SecretConn *p2pconn.SecretConnection

	// Reader is used to read from the TCP stream.
	Reader protoio.ReadCloser

	// Writer is used to write to the TCP stream.
	Writer protoio.WriteCloser
}

// NewReadWriteConn returns a new instance of ReadWriteConn.
func NewReadWriteConn() *ReadWriteConn {
	return &ReadWriteConn{
		SecretConn: new(p2pconn.SecretConnection),
	}
}

// RetrySecretDial dials the given address until success and returns
// a secret connection.
func RetrySecretDial(protocol, address string, privkey ed25519.PrivateKey, logger *log.Logger) (*p2pconn.SecretConnection, error) {
	logger.Printf("[INFO] signctrl: Dialing validator at %v...\n", address)

	var conn net.Conn
	var err error

	for {
		if conn, err = net.Dial(protocol, address); err == nil {
			logger.Println("[INFO] signctrl: Successfully dialed validator. ✓")
			break
		}
		<-time.After(500 * time.Millisecond)
	}

	return p2pconn.MakeSecretConnection(conn, tmcrypto.PrivKey(privkey))
}
