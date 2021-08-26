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

package sim_test

import (
	"testing"
	"time"

	"github.com/go-air/reach/sim"

	"github.com/go-air/gini/logic"
	"github.com/go-air/gini/z"
)

const (
	Log = false
)

func TestSimEvents(t *testing.T) {
	trans := logic.NewS()
	N := 10
	ms := make([]z.Lit, N)
	carry := trans.T
	ins := make([]z.Lit, 8)
	all := trans.T
	for i := range ins {
		ins[i] = trans.Lit()
		all = trans.And(all, ins[i])
	}
	allm := trans.T
	for i := range ms {
		m := trans.Latch(trans.F)
		ms[i] = m
		//trans.SetNext(m, trans.Choice(all, trans.Choice(carry, m.Not(), m), m))
		allm = trans.And(allm, m)
	}
	for i := range ms {
		m := ms[i]
		c := trans.And(all, carry)
		trans.SetNext(m, trans.Choice(allm, m.Not(), trans.Choice(c, m.Not(), m)))
		carry = trans.And(carry, m)
	}
	s := sim.New(trans, carry)
	opts := sim.NewOptions()
	opts.WatchUntil = 100
	opts.EventChan = make(chan *sim.Event)
	opts.EventFlags = sim.FlagRoundTrip | sim.FlagWait
	opts.Duration = time.Hour
	s.SetOptions(opts)
	go s.Simulate()
	for i := 0; i < 10; i++ {
		ev := <-opts.EventChan
		if Log {
			t.Logf("got event %s\n", ev)
		}
		for _, m := range trans.Latches {
			v := ev.V[m.Var()]&(1<<uint(ev.I)) != 0
			if Log {
				t.Logf("\t%s: %t\n", m, v)
			}
		}
		opts.EventChan <- ev
	}
	ev := <-opts.EventChan
	ev.F = sim.FlagStop
	opts.EventChan <- ev
	ev, ok := <-opts.EventChan
	if ok {
		t.Errorf("didn't close after stop")
	}
}
