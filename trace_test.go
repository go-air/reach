// Copyright 2018 Scott Cotton. All rights reserved.  Use of this source
// code is governed by a license that can be found in the License file.

package reach

import (
	"bytes"
	"io/ioutil"
	"math/rand"
	"os"
	"testing"

	"github.com/go-air/gini/logic"
	"github.com/go-air/gini/z"
)

func TestTraceAppend(t *testing.T) {
	s, n, carry, ms := gen()
	tr := NewTrace(s, carry)
	all := make([][]bool, 1<<3)
	_ = n
	_ = ms
	for d := 0; d < 1<<3; d++ {
		vs := make([]bool, s.Len())
		if rand.Intn(2) == 1 {
			vs[n.Var()] = true
		}
		s.Eval(vs)
		tr.Append(vs)
		all[d] = vs
	}
	for d := 0; d < 1<<3; d++ {
		vs := all[d]
		for i, m := range tr.Latches {
			if tr.LatchVal(i, d) != vs[m.Var()] {
				t.Errorf("LatchVal depth %d, latch %d\n", d, i)
			}
		}
	}
}

func gen() (*logic.S, z.Lit, z.Lit, []z.Lit) {
	s := logic.NewS()
	n := s.Lit()
	bs := make([]z.Lit, 3)
	carry := n
	for i := range bs {
		bs[i] = s.Latch(s.F)
	}
	for _, m := range bs {
		s.SetNext(m, s.Choice(carry, m.Not(), m))
		carry = s.And(carry, m)
	}
	return s, n, carry, bs
}

func TestTraceM(t *testing.T) {
	s := logic.NewS()
	m := s.Latch(s.F)
	n := s.Lit()
	s.SetNext(m, s.Choice(n, m.Not(), m))
	tr := NewTrace(s)
	vsA, vsB := make([]bool, s.Len()), make([]bool, s.Len())
	for d := 0; d < 16; d++ {
		if d&1 != 0 {
			vsA[n.Var()] = true
		}
		s.Eval(vsA)
		tr.Append(vsA)
		nxt := s.Next(m)
		t := vsA[nxt.Var()]
		if !nxt.IsPos() {
			t = !t
		}
		vsB[m.Var()] = t
		vsA, vsB = vsB, vsA
	}
	if err := tr.Verify(s); err != nil {
		t.Error(err)
	}
}

func TestTraceVerifyValid(t *testing.T) {
	s, n, carry, ms := gen()
	tr := NewTrace(s, carry)
	vsA, vsB := make([]bool, s.Len()), make([]bool, s.Len())
	for i := 0; i < 1<<3; i++ {
		vsA[n.Var()] = true
		s.Eval(vsA)
		tr.Append(vsA)
		for _, m := range ms {
			nxt := s.Next(m)
			t := vsA[nxt.Var()]
			if !nxt.IsPos() {
				t = !t
			}
			vsB[m.Var()] = t
		}
		vsA, vsB = vsB, vsA
	}
	if err := tr.Verify(s); err != nil {
		t.Error(err)
	}
}

func TestTraceVerifyInvalidS(t *testing.T) {
	s, n, carry, ms := gen()
	tr := NewTrace(s, carry)
	vsA, vsB := make([]bool, s.Len()), make([]bool, s.Len())
	for i := 0; i < 1<<3; i++ {
		vsA[n.Var()] = true
		s.Eval(vsA)
		tr.Append(vsA)
		for _, m := range ms {
			nxt := s.Next(m)
			t := vsA[nxt.Var()]
			if !nxt.IsPos() {
				t = !t
			}
			vsB[m.Var()] = t
		}
		vsA, vsB = vsB, vsA
	}
	s.SetNext(ms[2], s.T)
	err := tr.Verify(s)
	if err == nil {
		t.Errorf("verified invalid init")
	} else {
		t.Logf("correctly found err %s", err)
	}
}

func TestTraceVerifyInvalidInit(t *testing.T) {
	s, n, carry, ms := gen()
	tr := NewTrace(s, carry)
	vsA, vsB := make([]bool, s.Len()), make([]bool, s.Len())
	vsA[ms[1].Var()] = true // violate init (Eval doesn't see it...)
	for i := 0; i < 1<<3; i++ {
		vsA[n.Var()] = true
		s.Eval(vsA)
		tr.Append(vsA)
		for _, m := range ms {
			nxt := s.Next(m)
			t := vsA[nxt.Var()]
			if !nxt.IsPos() {
				t = !t
			}
			vsB[m.Var()] = t
		}
		vsA, vsB = vsB, vsA
	}
	err := tr.Verify(s)
	if err == nil {
		t.Errorf("verified invalid init")
	} else {
		t.Logf("correctly found err %s", err)
	}
}

// this works in bytes.Buffer but not file...., but yes in buffered input from file...
var weirdBug = "trace 3 7 49 1\n2 3 4 5 6 7 8\n9 10 11 12 13 14 15 16 17 18 19 20 21 22 23 24 25 26 27 28 29 30 31 32 33 34 35 36 37 38 39 40 41 42 43 44 45 46 47 48 49 50 51 52 53 54 55 56 57\n809\n\x00\x00\x00\x00\x00\x00\x00·\x00d3\x00C\x00\x00\x00\x00 "

func TestTraceWeird(t *testing.T) {
	r := bytes.NewBuffer([]byte(weirdBug))
	if _, err := DecodeTrace(r); err != nil {
		t.Error(err)
	}
	tmp, err := ioutil.TempFile("", "weird")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write([]byte(weirdBug)); err != nil {
		t.Fatal(err)
	}
	if err := tmp.Close(); err != nil {
		t.Fatal(err)
	}
	f, err := os.Open(tmp.Name())
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if _, err := DecodeTrace(f); err != nil {
		t.Fatal(err)
	}
}

func TestTraceIO(t *testing.T) {
	trace := &Trace{}
	trace.n = 123
	trace.Inputs = make([]z.Lit, 27)
	trace.Latches = make([]z.Lit, 48)
	trace.Watches = make([]z.Lit, 3)
	for i := range trace.Inputs {
		trace.Inputs[i] = z.Lit(rand.Intn(96) + 2).Var().Pos()
	}
	for i := range trace.Latches {
		trace.Latches[i] = z.Lit(rand.Intn(23) + 127).Var().Pos()
	}
	for i := range trace.Watches {
		trace.Watches[i] = z.Lit(rand.Intn(5) + 511)
	}
	sz := len(trace.Inputs) + len(trace.Watches) + len(trace.Latches)
	nv := trace.n * sz
	trace.values = make([]bool, nv)
	for i := 0; i < nv; i++ {
		if rand.Intn(2) == 1 {
			trace.values[i] = true
		}
	}
	w := bytes.NewBuffer(nil)
	if err := trace.Encode(w); err != nil {
		t.Fatal(err)
	}
	r := bytes.NewBuffer(w.Bytes())
	ttrace, err := DecodeTrace(r)
	if err != nil {
		t.Fatal(err)
	}
	if ttrace.n != trace.n {
		t.Errorf("N")
	}
	if len(trace.values) != len(ttrace.values) {
		t.Errorf("|vals|")
	}
	for i, v := range trace.values {
		if ttrace.values[i] != v {
			t.Errorf("val %d: got %t not %t\n", i, ttrace.values[i], v)
		}
	}
	if len(trace.Inputs) != len(ttrace.Inputs) {
		t.Errorf("|ins|")
	}
	for i, m := range trace.Inputs {
		if ttrace.Inputs[i] != m {
			t.Errorf("input %d got %s not %s\n", i, ttrace.Inputs[i], m)
		}
	}
	if len(trace.Latches) != len(ttrace.Latches) {
		t.Errorf("|latches|")
	}
	for i, m := range trace.Latches {
		if ttrace.Latches[i] != m {
			t.Errorf("latch %d got %s not %s\n", i, ttrace.Latches[i], m)
		}
	}
	if len(trace.Watches) != len(ttrace.Watches) {
		t.Errorf("|watches|")
	}
	for i, m := range trace.Watches {
		if ttrace.Watches[i] != m {
			t.Errorf("watch %d got %s not %s\n", i, ttrace.Watches[i], m)
		}
	}
}
