// Copyright (c) 2021 The Reach authors (see AUTHORS)
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of
// this software and associated documentation files (the "Software"), to deal in
// the Software without restriction, including without limitation the rights to
// use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
// the Software, and to permit persons to whom the Software is furnished to do so,
// subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
// FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
// COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
// CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

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
