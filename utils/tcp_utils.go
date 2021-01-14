package utils

import (
	"crypto/ed25519"
	"net"

	tmcrypto "github.com/tendermint/tendermint/crypto/ed25519"
	p2pconn "github.com/tendermint/tendermint/p2p/conn"
)

// RetrySecretDial dials the given address until success and returns
// a secret connection.
func RetrySecretDial(protocol, address string, privkey ed25519.PrivateKey) (*p2pconn.SecretConnection, error) {
	var conn net.Conn
	var err error

	for {
		if conn, err = net.Dial(protocol, address); err == nil {
			break
		}
	}

	secretConn, err := p2pconn.MakeSecretConnection(conn, tmcrypto.PrivKey(privkey))
	if err != nil {
		return nil, err
	}

	return secretConn, nil
}
