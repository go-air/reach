// Copyright 2018 Scott Cotton. All rights reserved.  Use of this source
// code is governed by a license that can be found in the License file.

package reach

import (
	"github.com/go-air/gini/logic"
	"github.com/go-air/gini/z"
)

// Primer takes literals in a logic.S and returns the corresponding literal with
// latches replaced with their next states.
type Primer struct {
	trans  *logic.S
	primed []z.Lit
}

// NewPrimer creates a new primer for a sequential system in type *logic.S for
// the latches in `t` and the properties specified in `ps`.
//
// NewPrimer may, and usually does, add nodes `t`.
func NewPrimer(t *logic.S, ps ...z.Lit) *Primer {
	primed := make([]z.Lit, t.Len())
	res := &Primer{trans: t, primed: primed}
	for _, m := range t.Latches {
		primeRec(t, m, primed)
	}
	for _, p := range ps {
		primeRec(t, p, primed)
	}
	return res
}

// Prime finds the primed version of a literal `m` in the transition system
// `trans` passed to NewPrimer. `m` should have been present in `trans` when
// NewPrimer was called. If this is not the case, Prime may panic or otherwise
// have undefined behavior.
func (p *Primer) Prime(m z.Lit) z.Lit {
	mv := m.Var()
	mvp := p.primed[mv]
	if m.IsPos() {
		return mvp
	}
	return mvp.Not()
}

func primeRec(trans *logic.S, p z.Lit, primed []z.Lit) z.Lit {
	pVar := p.Var()
	prime := primed[pVar]
	if prime != z.LitNull {
		if p.IsPos() {
			return prime
		}
		return prime.Not()
	}
	switch trans.Type(p) {
	case logic.SAnd:
		a, b := trans.Ins(p)
		a, b = primeRec(trans, a, primed), primeRec(trans, b, primed)
		res := trans.And(a, b)

		primed[pVar] = res
		if p.IsPos() {
			return res
		}
		return res.Not()
	case logic.SInput:
		primed[pVar] = pVar.Pos()
		return p
	case logic.SConst:
		primed[pVar] = pVar.Pos()
		return p
	case logic.SLatch:
		res := trans.Next(p)
		primed[pVar] = res
		if p.IsPos() {
			return res
		}
		return res.Not()
	default:
		panic("unreachable")
	}
}
