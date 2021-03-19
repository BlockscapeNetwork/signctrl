package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRegisterGauges(t *testing.T) {
	g := RegisterGauges()
	assert.NotNil(t, g.RankGauge)
	assert.NotNil(t, g.MissedInARowGauge)
}
