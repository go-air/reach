// Copyright 2018 The Reach Authors. All rights reserved.  Use of this source
// code is governed by a license that can be found in the License file.

package iic

import (
	"testing"

	"github.com/go-air/gini"
	"github.com/go-air/gini/logic"
	"github.com/go-air/gini/z"
	"github.com/go-air/reach"
)

func TestJustifyAndOrInLatch(t *testing.T) {
	trans := logic.NewS()
	ms := make([]z.Lit, 10)
	ns := make([]z.Lit, 10)
	for i := range ms {
		ms[i] = trans.Latch(z.LitNull)
		ns[i] = trans.Lit()
	}
	prop := trans.F
	for i := range ms {
		prop = trans.Or(prop, ms[i])
		prop = trans.Or(prop, ns[i])
	}

	just := newJustifier(trans)
	sat := gini.New()
	trans.ToCnf(sat)
	sat.Assume(prop)
	if sat.Solve() != 1 {
		t.Fatal("unsat1")
	}

	o := just.Justify(nil, sat, prop)
	sat.Assume(prop.Not())
	if sat.Solve() != 1 {
		t.Fatal("unsat2")
	}
	a := just.Justify(nil, sat, prop.Not())
	if len(o) != 1 {
		t.Errorf("or: %d", len(o))
	}
	if len(a) != len(ms) {
		t.Errorf("too few for and %d", len(a))
	}
}

func TestJustifyBadPrimeCounter(t *testing.T) {
	N := 3
	trans := logic.NewS()
	ms := make([]z.Lit, N)
	in := trans.Lit()
	carry := trans.T
	init := trans.T
	for i := range ms {
		m := trans.Latch(trans.F)
		ms[i] = m
		trans.SetNext(m, trans.Choice(trans.And(in, carry), m.Not(), m))
		carry = trans.And(carry, m)
		init = trans.And(init, m.Not())
	}
	sat := gini.New()
	primer := reach.NewPrimer(trans, init, carry)
	trans.ToCnf(sat)
	just := newJustifier(trans)
	carryPrime := primer.Prime(carry)
	sat.Assume(carryPrime)
	if sat.Solve() != 1 {
		t.Fatal("unsat")
	}
	js := just.Justify(nil, sat, carryPrime)
	if len(js) != N {
		t.Errorf("didn't justify: %v => %v\n", js, ms)
	}
}
