package sim

import "time"

// Options provides configuration info
// for the simulutor.
type Options struct {
	// MaxDepth to which to simulate (unless RestartFactor is non-zero).
	MaxDepth int64
	// The max duration of a simulation.
	Duration time.Duration
	// Stop if every watch has been positive WatchUntil times.
	WatchUntil int
	// The random Seed
	Seed int64
	// N is the number of simulations to run, default 1.
	N int
	// TraceWindow give the max size of a trace (window of simulation memory), default 128.
	TraceWindow int
	// RestartFactor, given a function k=f(n) telling us to restart the n'th time
	// after k steps, run RestartFactor*k steps.  f is the Luby Series.
	RestartFactor int
	// GenTrace whether to generate a trace.
	GenTrace bool
	// log events
	Verbose bool
	// Events, ignored if EventChan is nil
	EventFlags EventFlag
	// EventChan is a channel on which to communicate simulation events.
	EventChan chan *Event
}

// NewOptions gives default options.
func NewOptions() *Options {
	return &Options{
		MaxDepth:      1 << 30,
		Duration:      time.Second,
		WatchUntil:    1,
		Seed:          44,
		N:             1,
		TraceWindow:   128,
		RestartFactor: 0,
		GenTrace:      true}
}
