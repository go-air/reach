package iic

import "time"

// Options for the IIC checker
type Options struct {
	Verbose         bool
	Preprocess      bool
	MaxDepth        int
	Duration        time.Duration
	GenTrace        bool
	VerifyInvariant bool
	FilterObs       bool
	ConsecuSift     bool
	ConsecuSiftPull bool
	Justify         bool
	DeepObs         bool
	GnrlRemoveLits  bool
}

// NewOptions gives a new Options object with
// default values.  The zero value is not default.
func NewOptions() *Options {
	return &Options{
		Verbose:         false,
		Preprocess:      true,
		MaxDepth:        1 << 30,
		Duration:        time.Second,
		GenTrace:        true,
		VerifyInvariant: true,
		FilterObs:       true,
		ConsecuSift:     true,
		ConsecuSiftPull: true,
		Justify:         true,
		DeepObs:         true,
		GnrlRemoveLits:  false}
}
