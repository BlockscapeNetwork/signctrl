package types

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Gauges wraps SignCTRL's prometheus gauges.
type Gauges struct {
	RankGauge         prometheus.Gauge
	MissedInARowGauge prometheus.Gauge
}

// RegisterGauges registers SignCTRL's prometheus gauges and returns them.
func RegisterGauges() Gauges {
	var g Gauges
	g.RankGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "signctrl_rank",
		Help: "Current rank of the SignCTRL validator.",
	})
	g.MissedInARowGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "signctrl_missed_blocks_in_a_row",
		Help: "Number of blocks missed in a row",
	})

	return g
}
