// Copyright 2018 The Reach Authors. All rights reserved.  Use of this source
// code is governed by a license that can be found in the License file.

package iic

import (
	"testing"
	"time"

	"github.com/go-air/gini"
	"github.com/go-air/gini/logic"
	"github.com/go-air/gini/z"
	"github.com/go-air/reach"
)

func TestCounterSat(t *testing.T) {
	N := 4
	trans := logic.NewS()
	ms := make([]z.Lit, N)
	in := trans.Lit()
	carry := trans.T
	init := trans.T
	bad := trans.T
	for i := range ms {
		m := trans.Latch(trans.F)
		ms[i] = m
		trans.SetNext(m, trans.Choice(trans.And(carry, in), m.Not(), m))
		carry = trans.And(carry, m)
		init = trans.And(init, m.Not())
		bad = trans.And(bad, m)
	}
	dead := time.Now().Add(time.Hour)
	sat := newSatMon("test", gini.New(), &dead)

	pri := reach.NewPrimer(trans, init, bad)
	trans.ToCnf(sat.sat)
	for i := range ms {
		mp := pri.Prime(ms[i])
		sat.Assume(mp)
	}
	sat.Assume(init)
	res := sat.Try()
	if res == 1 {
		t.Errorf("sat")
	}
}

func TestIicTrivInd(t *testing.T) {
	trans := logic.NewS()
	m := trans.Latch(trans.F)
	trans.SetNext(m, m)
	mc := New(trans, m)
	if mc.Try() != -1 {
		t.Errorf("got trace in triv unsat.")
	}
}

func TestIicCounter(t *testing.T) {
	N := 10
	trans := logic.NewS()
	ms := make([]z.Lit, N)
	in := trans.Lit()
	carry := trans.T
	for i := range ms {
		m := trans.Latch(trans.F)
		ms[i] = m
		trans.SetNext(m, trans.Choice(trans.And(carry, in), m.Not(), m))
		carry = trans.And(carry, m)
	}
	mc := New(trans, carry)
	switch mc.Try() {
	case 1:
	case 0:
		t.Logf("iic timed out")
	case -1:
		t.Errorf("got ind, expected cex")
	}
}

func TestIicNotCounter(t *testing.T) {
	N := 3
	trans := logic.NewS()
	ms := make([]z.Lit, N)
	in := trans.Lit()
	carry := trans.T
	for i := range ms {
		m := trans.Latch(trans.F)
		ms[i] = m
		trans.SetNext(m, trans.Choice(trans.And(carry, in), m.Not(), m))
		carry = trans.And(carry, m)
	}
	trans.SetNext(ms[N-1], trans.And(trans.Next(ms[N-1]), ms[0].Not()))
	mc := New(trans, carry)
	switch mc.Try() {
	case -1:
	case 0:
		t.Logf("timed out...")
	case 1:
		t.Errorf("got cex, expected inv")
	}
}

func TestIicFifo(t *testing.T) {
	N := 4
	trans := logic.NewS()
	advance := trans.Lit()
	last := trans.Lit()
	all := trans.T
	ms := make([]z.Lit, N)
	for i := range ms {
		m := trans.Latch(trans.F)
		trans.SetNext(m, trans.Choice(advance, last, m))
		ms[i] = m
		last = m
		all = trans.And(all, m)
	}
	mc := New(trans, all)
	switch mc.Try() {
	case 1:
	case 0:
		t.Logf("timed out.\n")
	case -1:
		t.Errorf("got ind, expected cex")
	}
}
