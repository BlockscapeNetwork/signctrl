package privval

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetStatus(t *testing.T) {
	pv := mockSCFilePV(t)
	err := pv.StartHTTPServer()
	assert.NoError(t, err)

	sr, err := GetStatus()
	assert.NotNil(t, sr)
	assert.NoError(t, err)
}
