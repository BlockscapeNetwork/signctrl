package connection

import (
	"crypto/ed25519"
	"crypto/rand"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"sync"
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

func TestRetryDialTCP_NoConnKey(t *testing.T) {
	cfgDir := "./test_dial_tcp_withconnkey"
	os.MkdirAll(cfgDir, 0700)
	defer os.RemoveAll(cfgDir)

	port, _ := getFreePort(t)
	laddr := fmt.Sprintf("127.0.0.1:%v", port)
	_, priv, _ := ed25519.GenerateKey(rand.Reader)
	go startMockTCPServer(t, laddr, priv, 0)

	conn, err := RetryDial(cfgDir, "tcp://"+laddr, log.New(ioutil.Discard, "", 0))
	assert.Nil(t, conn)
	assert.Error(t, err)
}

func TestRetryDialTCP_WithConnKey(t *testing.T) {
	cfgDir := "./test_dial_tcp_withconnkey"
	os.MkdirAll(cfgDir, 0700)
	defer os.RemoveAll(cfgDir)

	err := CreateBase64ConnKey(cfgDir)
	assert.NoError(t, err)

	port, _ := getFreePort(t)
	laddr := fmt.Sprintf("127.0.0.1:%v", port)
	_, priv, _ := ed25519.GenerateKey(rand.Reader)
	go startMockTCPServer(t, laddr, priv, 1100*time.Millisecond)

	conn, err := RetryDial(cfgDir, "tcp://"+laddr, log.New(ioutil.Discard, "", 0))
	assert.NotNil(t, conn)
	assert.NoError(t, err)
}

func startMockUnixServer(t *testing.T, laddr string, delay time.Duration, wg *sync.WaitGroup) error {
	t.Helper()
	time.Sleep(delay)
	defer wg.Done()

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
	cfgDir := "./test_dial_unix"
	sockAddr := fmt.Sprintf("%v/test.sock", cfgDir)
	os.MkdirAll(cfgDir, 0700)
	defer os.RemoveAll(cfgDir)

	var wg sync.WaitGroup
	wg.Add(1)
	go startMockUnixServer(t, sockAddr, 1100*time.Millisecond, &wg)

	conn, err := RetryDial(cfgDir, "unix://"+sockAddr, log.New(ioutil.Discard, "", 0))
	assert.NotNil(t, conn)
	assert.NoError(t, err)

	wg.Wait()
}

func TestRetryDialUnknown(t *testing.T) {
	conn, err := RetryDial(".", "invalid://127.0.0.1:3000", log.New(ioutil.Discard, "", 0))
	assert.Nil(t, conn)
	assert.Error(t, err)
}
