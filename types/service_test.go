package types

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type testService struct {
	BaseService
}

func TestStartStopService(t *testing.T) {
	ts := &testService{}
	ts.BaseService = *NewBaseService(nil, "TestService", ts)

	name := ts.String()
	assert.Equal(t, "TestService", name)

	err := ts.Start()
	assert.NoError(t, err)

	running := ts.IsRunning()
	assert.Equal(t, true, running)

	err = ts.Start()
	assert.Error(t, err)

	err = ts.Stop()
	assert.NoError(t, err)

	running = ts.IsRunning()
	assert.Equal(t, false, running)

	err = ts.Stop()
	assert.Error(t, err)
}

func TestWait(t *testing.T) {
	ts := &testService{}
	ts.BaseService = *NewBaseService(nil, "TestService", ts)

	err := ts.Start()
	assert.NoError(t, err)

	waitCh := make(chan struct{})
	go func() {
		ts.Wait()
		waitCh <- struct{}{}
	}()

	go ts.Stop()

	select {
	case <-waitCh:
		break
	case <-time.After(100 * time.Millisecond):
		t.Fatal("expected Wait() to finish within 100ms")
	}
}

func TestQuit(t *testing.T) {
	ts := &testService{}
	ts.BaseService = *NewBaseService(nil, "TestService", ts)

	err := ts.Start()
	assert.NoError(t, err)

	go ts.Stop()

	select {
	case <-ts.Quit():
		break
	case <-time.After(100 * time.Millisecond):
		t.Fatal("expected Quit() to finish within 100ms")
	}
}
