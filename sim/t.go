// Copyright 2018 Scott Cotton. All rights reserved.  Use of this source
// code is governed by a license that can be found in the License file.

package sim

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/go-air/reach"

	"github.com/go-air/gini/logic"
	"github.com/go-air/gini/z"
)

// T holds state for a simulator.
type T struct {
	trans       *logic.S
	inputs      []z.Lit
	watches     []z.Lit
	traces      []*reach.Trace
	depths      []int64
	vsA, vsB    []uint64
	rnd         *rand.Rand
	deadLine    time.Time
	watchCounts [][64]int
	window      [][]uint64
	wi          int
	steps       int64
	luby        *luby

	opts *Options
}

// New creates a new simulator.
func New(trans *logic.S, bads ...z.Lit) *T {
	res := &T{trans: trans}
	ws := make([]z.Lit, len(bads))
	for i, b := range bads {
		ws[i] = b
	}
	res.watches = ws
	res.traces = make([]*reach.Trace, len(ws))
	res.depths = make([]int64, len(ws))
	for i := range res.depths {
		res.depths[i] = -1
	}
	res.vsA = make([]uint64, trans.Len())
	res.vsB = make([]uint64, trans.Len())
	res.watchCounts = make([][64]int, len(ws))
	for i := 2; i < trans.Len(); i++ {
		m := z.Var(i).Pos()
		if trans.Type(m) == logic.SInput {
			res.inputs = append(res.inputs, m)
		}
	}
	res.opts = NewOptions()
	res.SetOptions(res.opts)
	return res
}

func (t *T) SetOptions(opts *Options) {
	t.opts = opts
	t.setWindow(opts.TraceWindow)
	t.rnd = rand.New(rand.NewSource(t.opts.Seed))
	if opts.RestartFactor != 0 {
		t.luby = newLuby()
	}
}

// SetWindow sets the window or
// memory size for generating traces.
func (t *T) setWindow(n int) {
	N := t.trans.Len()
	d := make([]uint64, N*n)
	t.window = make([][]uint64, n)
	for i := range t.window {
		t.window[i] = d[N*i : N*(i+1)]
	}
}

// Simulate runs the simulation with the current options.
func (t *T) Simulate() int64 {
	ticker := time.NewTicker(time.Second)
	defer func() {
		ticker.Stop()
		if t.opts.EventChan != nil {
			close(t.opts.EventChan)
			t.opts.EventChan = nil
		}
	}()

	ttl := int64(0)

	for i := 0; i < t.opts.N; i++ {
		if t.opts.RestartFactor != 0 {
			t.opts.MaxDepth = int64(int(t.luby.Next()) * t.opts.RestartFactor)
		}
		ttl += t.simulateOne(ticker)
	}
	return ttl
}

// SimulateFor runs the simulation for at most
// `dur` time.  Simulation may stop early as per
// other configuration.
func (t *T) simulateOne(ticker *time.Ticker) int64 {
	t.deadLine = time.Now().Add(t.opts.Duration)
	res := int64(0)
	t.init()
	trans := t.trans
	if t.opts.Verbose {
		fmt.Printf("[sim] initialized ... starting simulation.\n")
	}
	for {
		if t.opts.Verbose {
			select {
			case <-ticker.C:
				fmt.Printf("[sim] step %d\n", t.steps)
			default:
			}
		}
		trans.Eval64(t.vsA)
		t.addStep(t.vsA)
		if time.Until(t.deadLine) <= 0 {
			if t.opts.Verbose {
				fmt.Printf("[sim] deadline reached after %d steps.\n", t.steps)
			}
			return res
		}
		if t.steps >= t.opts.MaxDepth {
			if t.opts.Verbose {
				fmt.Printf("[sim] maxdepth %d reached.\n", t.steps)
			}
			return res
		}
		if debugState {
			fmt.Printf("latch states:\n")
		}
		for i, m := range trans.Latches {
			mp := trans.Next(m)
			if debugState {
				vm := t.vsA[m.Var()]
				vp := t.vsA[mp.Var()]
				if !mp.IsPos() {
					vp = ^vp
				}
				fmt.Printf("\t%d. %s\n\t\t%b\n\t\t%b\n", i, m, vm, vp)
			}
			vp := t.vsA[mp.Var()]
			if !mp.IsPos() {
				vp = ^vp
			}
			t.vsB[m.Var()] = vp
		}
		for _, m := range t.inputs {
			t.vsB[m.Var()] = t.rnd.Uint64()
		}
		min := t.opts.WatchUntil
		for i, m := range t.watches {
			wvs := t.vsA[m.Var()]
			if !m.IsPos() {
				wvs = ^wvs
			}
			if wvs == 0 {
				min = 0
				continue
			}
			if debugState || t.opts.Verbose {
				fmt.Printf("[sim] watch %d: %s has %b\n", i, m, wvs)
			}
			ttl := 0
			for s := uint(0); s < 64; s++ {
				ttl += t.watchCounts[i][s]
				if (wvs & (1 << s)) == 0 {
					continue
				}
				if debugState {
					fmt.Printf("\t(index %d)\n", s)
				}
				ttl++
				t.watchCounts[i][s]++
				if t.depths[i] == -1 {
					t.depths[i] = t.steps
				}
				if t.traces[i] == nil {
					t.traces[i] = t.genTrace(m, s)
				}
				t.execEvent(m, int(s), t.traces[i])
			}
			if min > ttl {
				min = ttl
			}
		}
		t.vsA, t.vsB = t.vsB, t.vsA
		res++
		t.steps = res
		if min >= t.opts.WatchUntil {
			return res
		}
	}
}

func (t *T) fillEvent(m z.Lit, i int, tr *reach.Trace, ev *Event) {
	flag := t.opts.EventFlags
	ev.N = t.steps
	ev.M = m
	ev.I = i
	ev.F = flag
	ev.WI = t.wi
	if flag&FlagCopyV != 0 {
		ev.V = make([]uint64, t.trans.Len())
		copy(ev.V, t.vsA)
	} else {
		ev.V = t.vsA
	}
	if flag&FlagCopyW != 0 {
		ev.W = make([][]uint64, len(t.window))
		for i := range ev.W {
			ev.W[i] = make([]uint64, len(t.window[i]))
			copy(ev.W[i], t.window[i])
		}
	} else {
		ev.W = t.window
	}
	if flag&FlagTrace != 0 {
		ev.T = tr
	} else {
		ev.T = nil
	}
}

func (t *T) execEvent(m z.Lit, i int, tr *reach.Trace) bool {
	ch := t.opts.EventChan
	if ch == nil {
		return true
	}
	if t.opts.Verbose {
		fmt.Printf("[sim] executing watch event over channel.\n")
		defer fmt.Printf("[sim] done executing watch event.\n")
	}
	ev := &Event{}
	t.fillEvent(m, i, tr, ev)
	flag := t.opts.EventFlags
	if flag&FlagWait == 0 {
		select {
		case ch <- ev:
			if flag&FlagRoundTrip == 0 {
				return true
			}
			ev, ok := <-ch
			if !ok {
				return false
			}
			flag = ev.F
			if flag&FlagStop != 0 {
				close(ch)
				t.opts.EventChan = nil
				return false
			}
		default:
			return true
		}
	}
	ch <- ev
	if flag&FlagRoundTrip == 0 {
		return true
	}
	ev = <-ch
	t.opts.EventFlags = ev.F
	if ev.F&FlagStop != 0 {
		close(ch)
		t.opts.EventChan = nil
		return false
	}
	return true
}

// FillOutput fills `out` with the results of
// the last simulation.
func (t *T) FillOutput(out *reach.Output) {
	for i, w := range t.watches {
		b := &reach.Result{M: w}
		tr := t.traces[i]
		d := t.depths[i]
		if d != -1 {
			b.Depth = int(d) // TBD(wsc) overflow
			b.SetReachable(tr)
		}
		out.AppendResult(b)
	}
}

func (t *T) genTrace(w z.Lit, s uint) *reach.Trace {
	trace := reach.NewTrace(t.trans, w)
	vs := make([]bool, t.trans.Len())
	i := t.wi
	if i == t.wi {
		i = 0
	}
	if t.steps < int64(t.opts.TraceWindow) {
		i = 0
	}
	for {
		win := t.window[i]
		var v64 uint64
		for j := range vs {
			v64 = win[j]
			vs[j] = (v64 & (1 << s)) != 0
		}
		trace.Append(vs)
		i++
		if i > len(t.window) {
			i = 0
		}
		if i == t.wi {
			break
		}
	}
	return trace
}

func (t *T) addStep(vs []uint64) {
	copy(t.window[t.wi], vs)
	t.wi++
	if t.wi >= t.opts.TraceWindow {
		t.wi = 0
	}
}

func (t *T) init() {
	rnd := t.rnd
	for _, m := range t.inputs {
		t.vsA[m.Var()] = rnd.Uint64()
	}
	trans := t.trans
	for _, m := range trans.Latches {
		switch trans.Init(m) {
		case trans.T:
			t.vsA[m.Var()] = (1 << 64) - 1
		case trans.F:
			t.vsA[m.Var()] = 0
		default:
			t.vsA[m.Var()] = rnd.Uint64()
		}
	}
	t.steps = 0
}
