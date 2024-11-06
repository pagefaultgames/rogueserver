package savedata

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	gameModeCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "rogueserver_game_mode_total",
			Help: "The total number of classic sessions played per 5 minutes",
		},
		[]string{"gamemode"},
	)

	starterCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "rogueserver_starter_count",
			Help: "The total number of times a specific starter was selected",
		},
		[]string{"starterKey", "gameMode"},
	)
)
