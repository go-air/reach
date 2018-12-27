// Copyright 2018 Scott Cotton. All rights reserved.  Use of this source
// code is governed by a license that can be found in the License file.

package iic

import (
	"math/rand"

	"github.com/irifrance/gini/inter"
	"github.com/irifrance/gini/logic"
	"github.com/irifrance/gini/z"
)

// does circuit based justifiation minimizing number of latches.  Used as step
// 1 of 2 step generalization, where the second step expands the result (in
// terms of number of literals) to make sure the level CNF is also satisfied by
// all extensions of the partial assignment.
type justifier struct {
	trans          *logic.S
	marks          []int8
	latchInfluence []uint32
}

func newJustifier(trans *logic.S) *justifier {
	N := trans.Len()
	res := &justifier{trans: trans}
	res.marks = make([]int8, N)
	res.latchInfluence = make([]uint32, N)
	for v := z.Var(2); v < z.Var(N); v++ {
		res.latchInfluence[v] = res.latchIn(v.Pos())
	}
	return res
}

func (j *justifier) latchIn(m z.Lit) uint32 {
	if j.marks[m.Var()] == 1 {
		return j.latchInfluence[m.Var()]
	}
	j.marks[m.Var()] = 1
	switch j.trans.Type(m) {
	case logic.SLatch:
		j.latchInfluence[m.Var()] = 1
		return 1
	case logic.SInput:
		j.latchInfluence[m.Var()] = 0
		return 0
	case logic.SAnd:
		a, b := j.trans.Ins(m)
		resA := j.latchIn(a)
		resB := j.latchIn(b)
		res := resA + resB
		j.latchInfluence[m.Var()] = res
		return res
	case logic.SConst:
		j.latchInfluence[m.Var()] = 0
		return 0
	}
	panic("unreachable")
}

func (j *justifier) JustifyInit() {
	for i := range j.marks {
		j.marks[i] = 0
	}
}

func (j *justifier) JustifyOne(ms []z.Lit, model inter.Model, m z.Lit) []z.Lit {
	return j.justifyRec(ms, m, model)
}

func (j *justifier) Justify(dst []z.Lit, model inter.Model, ms ...z.Lit) []z.Lit {
	j.JustifyInit()
	for _, m := range ms {
		dst = j.justifyRec(dst, m, model)
	}
	return dst
}

func (j *justifier) justifyRec(dst []z.Lit, m z.Lit, model inter.Model) []z.Lit {
	mv := m.Var()
	mark := j.marks[mv]
	if mark == 1 {
		return dst
	}
	j.marks[mv] = 1
	if !model.Value(m) {
		m = m.Not()
	}
	switch j.trans.Type(m) {
	case logic.SLatch:
		dst = append(dst, m)
		return dst
	case logic.SInput, logic.SConst:
		return dst
	case logic.SAnd:
		a, b := j.trans.Ins(m)
		if m.IsPos() {
			dst = j.justifyRec(dst, a, model)
			dst = j.justifyRec(dst, b, model)
			return dst
		}
		if model.Value(a) {
			dst = j.justifyRec(dst, b, model)
			return dst
		}
		if model.Value(b) {
			dst = j.justifyRec(dst, a, model)
			return dst
		}
		if j.latchInfluence[a.Var()] > j.latchInfluence[b.Var()] {
			a, b = b, a
		} else if j.latchInfluence[a.Var()] == j.latchInfluence[b.Var()] {
			if rand.Intn(2) == 1 {
				a, b = b, a
			}
		}
		dst = j.justifyRec(dst, a, model)
		return dst
	}
	panic("unreachable")
}
