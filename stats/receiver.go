package stats

import "time"

type StatsReceiver interface {
	Counter(string) Counter
	Timer(string) Timer
	Scope(string) StatsReceiver
}

type Counter interface {
	Incr()
	Add(int64)
}

type Timer interface {
	Add(time.Duration)
}
