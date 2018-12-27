// Copyright 2018 Scott Cotton. All rights reserved.  Use of this source
// code is governed by a license that can be found in the License file.

package reach

import (
	"bufio"
	"fmt"
	"io"

	"github.com/irifrance/gini/inter"
	"github.com/irifrance/gini/logic"
	"github.com/irifrance/gini/z"
)

// Trace holds data for a trace of a sequential circuit as defined by
// `github.com/irifrance/gini/logic.S`, giving values to the latches, inputs,
// and an optional list of watched literals.
type Trace struct {
	n       int
	Inputs  []z.Lit
	Latches []z.Lit
	Watches []z.Lit
	values  []bool
}

// NewTrace creates a new trace of length 0 containing
// all the inputs and latches of `s` and the specified
// watches `ws`.
func NewTrace(s *logic.S, ws ...z.Lit) *Trace {
	return NewTraceLen(s, s.Len(), ws...)
}

// NewTraceLen creates a new trace of length 0 containing
// all the inputs and latches of `s` and the specified
// watches `ws` from the first `sLen` nodes of `s`.
//
// If `s` was created at one time suitable for making a
// trace, and had len `n==s.Len()` at this time, and later
// new latches or gates or inputs were defined in `s`, then
// the circuit `s` when it had len `n` may be used for
// constructing a new trace, specifying sLen as `n`.
func NewTraceLen(s *logic.S, sLen int, ws ...z.Lit) *Trace {
	N := sLen
	nIn, nL := 0, 0
	for i := 1; i < N; i++ {
		m := z.Var(i).Pos()
		switch s.Type(m) {
		case logic.SInput:
			nIn++
		case logic.SLatch:
			nL++
		}
	}
	latches := make([]z.Lit, nL)
	inputs := make([]z.Lit, nIn)
	li, ii := 0, 0
	for i := 1; i < N; i++ {
		m := z.Var(i).Pos()
		switch s.Type(m) {
		case logic.SInput:
			inputs[ii] = m
			ii++
		case logic.SLatch:
			latches[li] = m
			li++
		}
	}
	watches := make([]z.Lit, len(ws))
	copy(watches, ws)
	return &Trace{Inputs: inputs, Latches: latches, Watches: watches}
}

// NewTraceBmc creates a new trace given an unroller `u` and a model for the
// combinational circuit `u.C`.
//
// Since the unroller only contains cone of influence portion of a sequential
// circuit, the undefined values are taken by simulation.  The model is checked
// to be coherent with the simulation.  A non-nil error is returned iff there
// is incoherence.
func NewTraceBmc(u *logic.Roll, model inter.Model, ws ...z.Lit) (*Trace, []error) {
	res := NewTrace(u.S, ws...)
	N := u.MaxLen()
	vsA, vsB := make([]bool, u.S.Len()), make([]bool, u.S.Len())
	var t bool
	for _, m := range res.Latches {
		switch u.S.Init(m) {
		case u.S.T:
			t = true
		case u.S.F:
			t = false
		default:
			if u.Len(m) > 0 {
				t = model.Value(u.At(m, 0))
			} else {
				t = false
			}
			vsA[m.Var()] = t
		}
	}
	for d := 0; d < N; d++ {
		for _, m := range res.Inputs {
			if d >= u.Len(m) {
				// panic("wilma!")
				t = false
			} else {
				t = model.Value(u.At(m, d))
			}
			vsA[m.Var()] = t
		}
		u.S.Eval(vsA)
		res.Append(vsA)
		for _, m := range res.Latches {
			nxt := u.S.Next(m)
			t = vsA[nxt.Var()]
			if !nxt.IsPos() {
				t = !t
			}
			if d < u.Len(nxt) { //&& d < u.Len(m)-1 {
				if model.Value(u.At(nxt, d)) != t {
					// this actually can happen if an input not in the COI was set to false
					// needs investigation...
					fmt.Printf("depth %d latch %s nxt %s vsA[nxt]=%t C.T=%s At(nxt, %d)=%s len(nxt)=%d len(m)=%d\n", d, m,
						nxt, t, u.C.T, d, u.At(nxt, d), u.Len(nxt), u.Len(m))
				}
			}
			vsB[m.Var()] = t
		}
		vsA, vsB = vsB, vsA
	}
	if err := res.Verify(u.S); err != nil {
		return nil, err
	}
	return res, nil
}

// Append adds a new step to the trace.
//
// `vs` should be the truth values associated with an evaluation of a circuit
// with corresponding input, latch, and watch literals.  If `s` is such a
// circuit, then `len(vs) == s.Len()` and `vs` contains truth values for all
// gates in the circuit indexed by variable, as in
// `github.com/irifrance/gini/logic.C.Eval`
func (t *Trace) Append(vs []bool) {
	sz := len(t.Inputs) + len(t.Latches) + len(t.Watches)
	if cap(t.values) < len(t.values)+sz {
		m := (len(t.values) + sz) * 5 / 3
		if len(t.values) == 0 {
			m = sz * 13
		}
		tmp := make([]bool, len(t.values), m)
		copy(tmp, t.values)
		t.values = tmp
	}
	n := len(t.values)
	t.values = t.values[:n+sz]
	vals := t.values[n : n+sz]
	j := 0
	for _, m := range t.Inputs {
		vals[j] = vs[m.Var()]
		j++
	}
	for _, m := range t.Latches {
		vals[j] = vs[m.Var()]
		j++
	}
	for _, m := range t.Watches {
		t := vs[m.Var()]
		if !m.IsPos() {
			t = !t
		}
		vals[j] = t
		j++
	}
	t.n++
}

// Len returns the number of states in the trace.
func (t *Trace) Len() int {
	return t.n
}

// InputVal returns the truth value for input with index i at
// depth `depth`.
func (t *Trace) InputVal(i, depth int) bool {
	sz := len(t.Inputs) + len(t.Latches) + len(t.Watches)
	off := sz * depth
	return t.values[off+i]
}

// LatchVal returns the truth value for latch with index i at
// depth `depth`.
func (t *Trace) LatchVal(i, depth int) bool {
	sz := len(t.Inputs) + len(t.Latches) + len(t.Watches)
	off := sz*depth + len(t.Inputs)
	return t.values[off+i]
}

// WatchVal returns the truth value for the variable underlying the
// watched literal with index i at depth `depth`.
func (t *Trace) WatchVal(i, depth int) bool {
	sz := len(t.Inputs) + len(t.Latches) + len(t.Watches)
	off := sz*depth + len(t.Inputs) + len(t.Latches)
	return t.values[off+i]
}

// Verify verifies that the trace is coherent with simulation
// under `s` and that the simulation leads to every literal
// in t.Watches being true at some point.
//
// Verify returns a non nil error describing a latch or watch in a bad
// state in the trace with respect to s iff there is such
// a latch or watch.
//
// Verify may panic if the trace is not dimensioned according to `s`.
// Namely, if the latches in `t` are not latches in `s`, or likewise
// inputs.  For watches in `t`, the corresponding literals should
// exist in `s`.
func (t *Trace) Verify(s *logic.S) []error {
	var errors []error
	tv, err := newTv(t, s)
	if err != nil {
		errors = append(errors, err)
		return errors
	}
	N := t.n
	for i := 1; i < N; i++ {
		if err := tv.step(i); err != nil {
			errors = append(errors, err)
		}
	}
	s.Eval(tv.vsA)
	for i, m := range t.Watches {
		if tv.wCounts[i] != 0 {
			continue
		}
		wv := tv.vsA[m.Var()]
		if !m.IsPos() {
			wv = !wv
		}
		if !wv {
			errors = append(errors, fmt.Errorf("ErrWatchFalse: %s", m))
		}
	}
	return errors
}

// tv is a mini-simulator for verifying traces
// w.r.t. a concrete circuit.
type tv struct {
	trace    *Trace
	s        *logic.S
	vsA, vsB []bool
	wCounts  []int
}

func newTv(t *Trace, s *logic.S) (*tv, error) {
	N := s.Len()
	res := &tv{
		trace:   t,
		s:       s,
		vsA:     make([]bool, N),
		vsB:     make([]bool, N),
		wCounts: make([]int, len(t.Watches))}

	if N == 0 {
		return res, nil
	}
	// check dimensions
	if len(t.Latches) != len(s.Latches) {
		return nil, fmt.Errorf("ErrLatchCount: got %d not %d", len(s.Latches), len(t.Latches))
	}
	for _, m := range t.Latches {
		if s.Type(m) != logic.SLatch {
			return nil, fmt.Errorf("ErrNotLatch: %s %s", m, s.Type(m))
		}
	}
	j := 0
	V := z.Var(s.Len())
	for v := z.Var(2); v < V; v++ {
		m := v.Pos()
		if s.Type(m) != logic.SInput {
			continue
		}
		if j >= len(t.Inputs) {
			return nil, fmt.Errorf("ErrInputCount: too many inputs for trace: %d > %d",
				j+1, j)
		}
		j++
	}
	if j < len(t.Inputs) {
		return nil, fmt.Errorf("ErrInputCount: not enough inputs for trace: %d < %d",
			j, len(t.Inputs))
	}

	// check initial conditions and set latch variables in vsA.
	vs := res.vsA
	for i := range t.Latches {
		m := t.Latches[i]
		t := t.LatchVal(i, 0)
		switch s.Init(m) {
		case s.T:
			if !t {
				return nil, fmt.Errorf("latch %s set to %t but initialised to %t\n", m, t, true)
			}
		case s.F:
			if t {
				return nil, fmt.Errorf("latch %s set to %t but initialised to %t\n", m, t, false)
			}
		}
		vs[m.Var()] = t
	}
	// set inputs
	for i := range t.Inputs {
		m := t.Inputs[i]
		vs[m.Var()] = t.InputVal(i, 0)
	}
	// eval
	res.s.Eval(vs)
	// check watches
	for i, m := range t.Watches {
		mval := t.WatchVal(i, 0)
		if m.IsPos() == mval {
			res.wCounts[i]++
		}
		if !m.IsPos() {
			mval = !mval
		}
		if mval != vs[m.Var()] {
			if !m.IsPos() {
				return nil, fmt.Errorf("watch %s at 0 got %t not %t\n", m, !mval, !vs[m.Var()])
			}
			return nil, fmt.Errorf("watch %s at 0 got %t not %t\n", m, mval, vs[m.Var()])
		}
	}
	return res, nil
}

// only for d > 0
func (tv *tv) step(d int) error {
	trace := tv.trace
	s := tv.s
	vsA, vsB := tv.vsA, tv.vsB
	// latches
	for i, m := range trace.Latches {
		t := trace.LatchVal(i, d)
		nxt := s.Next(m)
		nv := vsA[nxt.Var()]
		if !nxt.IsPos() {
			nv = !nv
		}
		if t != nv {
			return fmt.Errorf("at %d latch %s (@%d) nxt %s got %t not %t\n", d, m, i, nxt, nv, t)
		}
		vsB[m.Var()] = t
	}
	// inputs
	for i := range trace.Inputs {
		m := trace.Inputs[i]
		t := trace.InputVal(i, d)
		vsB[m.Var()] = t
	}

	s.Eval(vsB)
	// watches
	for i, m := range trace.Watches {
		t := trace.WatchVal(i, d)
		if m.IsPos() == t {
			tv.wCounts[i]++
		}
		ref := vsB[m.Var()]
		if !m.IsPos() {
			ref = !ref
		}
		if ref != t {
			return fmt.Errorf("at %d, watch %s got %t not %t\n", d, m, ref, t)
		}
	}

	tv.vsA, tv.vsB = vsB, vsA
	return nil
}

// Encode writes a trace in a mostly binary format with some
// readable header info.
func (t *Trace) Encode(w io.Writer) error {
	var err error
	_, err = fmt.Fprintf(w, "trace %d %d %d %d\n", t.n, len(t.Inputs), len(t.Latches), len(t.Watches))
	if err != nil {
		return err
	}
	for _, ms := range [][]z.Lit{t.Inputs, t.Latches} {
		for i, m := range ms {
			if i == len(ms)-1 {
				_, err = fmt.Fprintf(w, "%d\n", m.Var())
			} else {
				_, err = fmt.Fprintf(w, "%d ", m.Var())
			}
			if err != nil {
				return err
			}
		}
	}
	for i, m := range t.Watches {
		if i < len(t.Watches)-1 {
			_, err = fmt.Fprintf(w, "%d ", m)
		} else {
			_, err = fmt.Fprintf(w, "%d\n", m)
		}
		if err != nil {
			return err
		}
	}
	chunkSz := len(t.Inputs) + len(t.Latches) + len(t.Watches)
	bSz := chunkSz / 8
	if chunkSz%8 != 0 {
		bSz++
	}
	buf := make([]byte, bSz)
	for i := 0; i < t.n; i++ {
		sl := t.values[i*chunkSz : i*chunkSz+chunkSz]
		var b byte
		bi := uint(0)
		buf = buf[:0]
		for _, t := range sl {
			if t {
				b |= 1 << bi
			}
			bi++
			if bi == 8 {
				buf = append(buf, b)
				bi = 0
				b = 0
			}
		}
		if bi != 0 {
			buf = append(buf, b)
		}
		_, err = w.Write(buf)
		if err != nil {
			return err
		}
	}
	return nil
}

// DecodeTrace tries to read a trace as written by Encode.
// DecodeTrace returns a non-nil error if there is an io or formatting
// error.
func DecodeTrace(r io.Reader) (*Trace, error) {
	r = bufio.NewReader(r)
	trace := &Trace{}
	var nIn, nL, nW int
	var err error
	_, err = fmt.Fscanf(r, "trace %d %d %d %d\n", &trace.n, &nIn, &nL, &nW)
	if err != nil {
		fmt.Printf("header\n")
		return nil, err
	}
	trace.Inputs = make([]z.Lit, nIn)
	trace.Latches = make([]z.Lit, nL)
	trace.Watches = make([]z.Lit, nW)
	var u uint32
	for i := range trace.Inputs {
		if i < nIn-1 {
			_, err = fmt.Fscanf(r, "%d ", &u)
		} else {
			_, err = fmt.Fscanf(r, "%d\n", &u)
		}
		if err != nil {
			fmt.Printf("input %d/%d\n", i, nIn)
			return nil, err
		}
		trace.Inputs[i] = z.Var(u).Pos()
	}
	for i := range trace.Latches {
		if i < len(trace.Latches)-1 {
			_, err = fmt.Fscanf(r, "%d ", &u)
		} else {
			_, err = fmt.Fscanf(r, "%d\n", &u)
		}
		if err != nil {
			fmt.Printf("latch %d\n", i)
			return nil, err
		}
		trace.Latches[i] = z.Var(u).Pos()
	}
	for i := range trace.Watches {
		if i < len(trace.Watches)-1 {
			_, err = fmt.Fscanf(r, "%d ", &u)
		} else {
			_, err = fmt.Fscanf(r, "%d\n", &u)
		}
		if err != nil {
			fmt.Printf("watch %d\n", i)
			return nil, err
		}
		trace.Watches[i] = z.Lit(u)
	}

	// read values now
	chunkSz := len(trace.Inputs) + len(trace.Latches) + len(trace.Watches)
	trace.values = make([]bool, chunkSz*trace.n)
	bSz := chunkSz / 8
	if chunkSz%8 != 0 {
		bSz++
	}
	buf := make([]byte, bSz)
	for i := 0; i < trace.n; i++ {
		ttl := 0
		for ttl < bSz {
			n, err := r.Read(buf[ttl:])
			if err != nil {
				return nil, err
			}
			ttl += n
		}
		tvs := trace.values[i*chunkSz : i*chunkSz+chunkSz]
		j := 0
		for _, b := range buf {
			for _, mask := range []byte{1, 2, 4, 8, 16, 32, 64, 128} {
				tvs[j] = b&mask != 0
				j++
				if j >= chunkSz {
					break
				}
			}
			if j >= chunkSz {
				break
			}
		}
	}
	return trace, nil
}

// EncodeAigerStim encodes a 'stimulus', which for an aiger file is just the
// sequence of inputs (since all initial values are assumed to be '0' in an
// aiger file.
//
// EncodeAigerStim returns the number of bytes written to dst and any error
// which occured in the process of encoding writing to dst.
func (t *Trace) EncodeAigerStim(dst io.Writer) (int, error) {
	sz := len(t.Inputs) + len(t.Latches) + len(t.Watches)
	N := len(t.values)
	nIn := len(t.Inputs)
	buf := make([]byte, nIn+1)
	vals := t.values
	ttl := 0
	var n int
	var err error
	for i := 0; i < N; i += sz {
		for j := range buf {
			if vals[i+j] {
				buf[j] = byte('1')
			} else {
				buf[j] = byte('0')
			}
		}
		buf[nIn] = byte('\n')
		n, err = dst.Write(buf)
		ttl += n
		if err != nil {
			return ttl, err
		}
	}
	buf[0] = byte('.')
	buf[1] = byte('\n')
	n, err = dst.Write(buf[:2])
	ttl += n
	return ttl, err
}
