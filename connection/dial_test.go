package connection

import (
	"crypto/ed25519"
	"crypto/rand"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	tm_ed25519 "github.com/tendermint/tendermint/crypto/ed25519"
	tm_p2pconn "github.com/tendermint/tendermint/p2p/conn"
)

// getFreePort asks the kernel for a free port that is ready to use.
func getFreePort(t *testing.T) (port int, err error) {
	t.Helper()
	var a *net.TCPAddr
	if a, err = net.ResolveTCPAddr("tcp", "localhost:0"); err == nil {
		var l *net.TCPListener
		if l, err = net.ListenTCP("tcp", a); err == nil {
			defer l.Close()
			return l.Addr().(*net.TCPAddr).Port, nil
		}
	}

	return
}

func startMockTCPServer(t *testing.T, laddr string, connKey ed25519.PrivateKey, delay time.Duration) error {
	t.Helper()
	time.Sleep(delay)

	listener, err := net.Listen("tcp", laddr)
	if err != nil {
		return err
	}
	defer listener.Close()

	conn, err := listener.Accept()
	if err != nil {
		return err
	}

	secretConn, err := tm_p2pconn.MakeSecretConnection(conn, tm_ed25519.PrivKey(connKey))
	if err != nil {
		return err
	}
	defer secretConn.Close()

	return nil
}

func TestRetryDialTCP(t *testing.T) {
	_, priv, _ := ed25519.GenerateKey(rand.Reader)

	// Fail loading conn.key.
	port, _ := getFreePort(t)
	laddr := fmt.Sprintf("127.0.0.1:%v", port)
	go startMockTCPServer(t, laddr, priv, 0)

	conn, err := RetryDial(".", "tcp://"+laddr, log.New(ioutil.Discard, "", 0))
	assert.Nil(t, conn)
	assert.Error(t, err)

	// Succeed loading conn.key.
	err = CreateBase64ConnKey(".")
	defer os.Remove("./conn.key")
	assert.NoError(t, err)

	port, _ = getFreePort(t)
	laddr = fmt.Sprintf("127.0.0.1:%v", port)
	go startMockTCPServer(t, laddr, priv, 1100*time.Millisecond)

	conn, err = RetryDial(".", "tcp://"+laddr, log.New(ioutil.Discard, "", 0))
	assert.NotNil(t, conn)
	assert.NoError(t, err)
}

func startMockUnixServer(t *testing.T, laddr string, delay time.Duration) error {
	t.Helper()
	time.Sleep(delay)

	listener, err := net.Listen("unix", laddr)
	if err != nil {
		return err
	}
	defer listener.Close()

	conn, err := listener.Accept()
	if err != nil {
		return err
	}
	defer conn.Close()

	return nil
}

func TestRetryDialUnix(t *testing.T) {
	go startMockUnixServer(t, "/tmp/test.sock", time.Second)
	defer os.Remove("/tmp/test.sock")
	conn, err := RetryDial(".", "unix:///tmp/test.sock", log.New(ioutil.Discard, "", 0))
	assert.NotNil(t, conn)
	assert.NoError(t, err)
}

func TestRetryDialUnknown(t *testing.T) {
	conn, err := RetryDial(".", "invalid://127.0.0.1:3000", log.New(ioutil.Discard, "", 0))
	assert.Nil(t, conn)
	assert.Error(t, err)
}
