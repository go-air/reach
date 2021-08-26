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

package sim

import (
	"fmt"

	"github.com/go-air/gini/z"
	"github.com/go-air/reach"
)

// Event gives the context of an event in a simulation.
type Event struct {
	M  z.Lit        // a watch
	I  int          // which trace in [0..64)
	V  []uint64     // evaluations
	W  [][]uint64   // window
	WI int          // index of next values in window.
	N  int64        // how many steps.
	T  *reach.Trace // trace or partial trace.
	F  EventFlag    // what flags were set.
}

// String implements Stringer.
func (ev *Event) String() string {
	return fmt.Sprintf("w: %s n: %d i: %d f: <%s>", ev.M, ev.N,
		ev.I, ev.F)
}

// EventFlag says what event info to send back
type EventFlag int

const (
	// FlagWait tells the simulator to continue if the listener is not ready to
	// receive.
	FlagWait EventFlag = 1 << iota

	// FlagRoundTrip tells the simulator to to ask for the event back after it
	// was sent.  When it receives the event, it updates the flags.  This can be
	// used to pause the simulation.
	FlagRoundTrip

	// FlagCopyV tells the simulator to copy the values.
	FlagCopyV

	// FlagCopyW tells the simulator to copy the window.
	FlagCopyW

	// FlagStop tells the simulator to stop.  It is only used in the case
	// FlagRoundTrip is set.
	FlagStop

	// FlagTrace tells the simulator to generate a trace for the watch.  Only
	// the first trace is generated.
	FlagTrace
)

func (f EventFlag) String() string {
	return fmt.Sprintf("%s|%s|%s|%s|%s|%s",
		nostr("wait", f&FlagWait != 0),
		nostr("roundtrip", f&FlagRoundTrip != 0),
		nostr("copyvals", f&FlagCopyV != 0),
		nostr("copywin", f&FlagCopyW != 0),
		nostr("stop", f&FlagStop != 0),
		nostr("trace", f&FlagTrace != 0))

}

func nostr(s string, v bool) string {
	if v {
		return s
	}
	return fmt.Sprintf("no%s", s)
}
