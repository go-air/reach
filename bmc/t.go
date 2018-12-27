// Copyright 2018 Scott Cotton. All rights reserved.  Use of this source
// code is governed by a license that can be found in the License file.

package bmc

import (
	"log"
	"time"

	"github.com/irifrance/gini"
	"github.com/irifrance/gini/logic"
	"github.com/irifrance/gini/z"

	"github.com/irifrance/reach"
)

type bmcBad struct {
	*reach.Result
	Timed     bool
	timeSpent time.Duration
}

func (b *bmcBad) remaining() time.Duration {
	return b.Dur - b.timeSpent
}

// T encapsulates a SAT based bounded model checker.
type T struct {
	sat  *gini.Gini
	roll *logic.Roll
	bads map[z.Lit]*bmcBad

	deadLine time.Time
	maxDepth int
	trace    bool
}

// New creates a new bounded model checker for bad states `bads` occuring in `s`.
//
// If `len(bads)==0`, then New panics.
func New(s *logic.S, bads ...z.Lit) *T {
	if len(bads) == 0 {
		panic("cannot do bmc without bad states!\n")
	}
	res := &T{sat: gini.NewVc(s.Len()*11, s.Len()*3*11), roll: logic.NewRoll(s)}
	res.bads = make(map[z.Lit]*bmcBad, len(bads))
	for _, m := range bads {
		res.bads[m] = &bmcBad{Result: &reach.Result{M: m}}
	}
	res.maxDepth = 1 << 30
	res.trace = true
	return res
}

// SetMaxDepth sets the maximum depth of subsequent bmc runs.
func (t *T) SetMaxDepth(d int) {
	t.maxDepth = d
}

// SetBadTimeout sets a timeout for the bad state `bad`.
//
// SetBadTimeout panics if `bad` was not supplied as a bad state in
// the call to `New()` that created `t`.
func (t *T) SetBadTimeout(bad z.Lit, d time.Duration) {
	b := t.bads[bad]
	b.Dur = d
}

// DistributeTimeout sets the timeout to `d`.  If
// there are `N` bad states in t, then each bad state
// has timeout set to `d / N`.
func (t *T) DistributeTimeout(d time.Duration) {
	badDur := d / time.Duration(len(t.bads))
	for _, b := range t.bads {
		b.Dur = badDur
	}
}

// Try tries to find paths to bad states within the constraints
// specified earlier (per bad timeout, max depth) and within
// duration `dur`.
//
// Run returns the number of reachable bad states found.
func (t *T) Try(dur time.Duration) int {
	t.deadLine = time.Now().Add(dur)
	found := 0
	depth := 0
	var mark []int8
	for found < len(t.bads) {
		if depth > t.maxDepth {
			return found
		}
		if time.Until(t.deadLine) <= 0 {
			return found
		}
		for k, v := range t.bads {
			if v.IsSolved() {
				continue
			}
			dur := time.Until(t.deadLine)
			if dur < 0 {
				return found
			}
			if v.Timed {
				if d := v.remaining(); d < dur {
					dur = d
				}
			}
			if dur < 0 {
				continue
			}
			start := time.Now()
			m := t.roll.At(k, depth)
			mark, _ = t.roll.C.CnfSince(t.sat, mark, m)
			t.sat.Assume(m)
			dur -= time.Since(start) // include unrolling time
			res := t.sat.Try(dur)
			if res == 1 {
				found++
				v.Status = 1
				v.Depth = depth
				if t.trace {
					tr, err := reach.NewTraceBmc(t.roll, t.sat, k)
					if err != nil {
						log.Fatal(err)
					}
					v.Trace = tr
				}
				continue
			}
			if v.Timed {
				v.timeSpent += time.Since(start)
			}
			if res == -1 {
				v.Depth = depth
			}
		}
		depth++
	}
	return found
}

// FillOutput fills the output object with
// bads and traces
func (t *T) FillOutput(dst *reach.Output) {
	for _, b := range t.bads {
		dst.AppendResult(b.Result)
	}
}
