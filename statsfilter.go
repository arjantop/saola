package saola

import (
	"time"

	"github.com/arjantop/saola/stats"
	"golang.org/x/net/context"
)

func NewStatsFilter(stats stats.StatsReceiver) Filter {
	requestsStat := stats.Counter("requests")
	successStat := stats.Counter("success")
	failureStat := stats.Counter("failure")
	latencyStat := stats.Timer("latency")
	return FuncFilter(func(ctx context.Context, s Service) error {
		start := time.Now()
		err := s.Do(ctx)
		latency := time.Now().Sub(start)

		requestsStat.Incr()
		latencyStat.Add(latency)
		if err != nil {
			failureStat.Incr()
		} else {
			successStat.Incr()
		}

		return err
	})
}
