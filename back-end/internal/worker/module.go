package worker

import (
	"mm/internal/worker/buyback"
	buybackpricetracker "mm/internal/worker/buyback_price_tracker"
	"mm/internal/worker/swaptarget"

	"go.uber.org/fx"
)

func Module() fx.Option {
	return fx.Module("worker",
		swaptarget.Module(),
		buybackpricetracker.Module(),
		buyback.Module(),
	)
}
