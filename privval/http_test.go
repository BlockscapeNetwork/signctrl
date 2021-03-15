package privval

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetStatus(t *testing.T) {
	pv := mockSCFilePV(t)
	pv.StartHTTPServer()

	sr, err := GetStatus()
	assert.NotNil(t, sr)
	assert.NoError(t, err)
}
