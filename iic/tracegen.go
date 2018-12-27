package iic

import (
	"fmt"

	"github.com/irifrance/gini/logic"
	"github.com/irifrance/gini/z"
	"github.com/irifrance/reach"
	"github.com/irifrance/reach/iic/internal/obs"
)

type traceGen struct {
	trans     *logic.S
	orgLen    int
	primer    *reach.Primer
	obs       *obs.Set
	init, bad z.Lit
	badPrime  z.Lit
	sat       *satmon
	hd        obs.Id
}

func (g *traceGen) build() (*reach.Trace, error) {
	trace := reach.NewTraceLen(g.trans, g.orgLen, g.bad)
	obA := g.hd
	vals := make([]bool, g.trans.Len()) // latch values
	// first set initial states.
	g.sat.Assume(g.init)
	obB := g.obs.Parent(obA)
	for obA != 0 {
		if debugTraceGen {
			fmt.Printf("doing trace step k=%d d=%d\n", g.obs.K(obA), g.obs.DistToBad(obA))
		}
		for _, m := range g.obs.Ms(obA) {
			if debugTraceGen {
				fmt.Printf("\tassuming %s from obA\n", m)
			}
			g.sat.Assume(m)
		}
		if g.obs.DistToBad(obA) == 1 {
			if debugTraceGen {
				fmt.Printf("\tassuming badprime: %s", g.badPrime)
			}
			g.sat.Assume(g.badPrime)
		} else if g.obs.DistToBad(obA) == 0 {
			if debugTraceGen {
				fmt.Printf("\tassuming bad: %s", g.bad)
			}
			g.sat.Assume(g.bad)
		}
		if obB != 0 {
			for _, m := range g.obs.Ms(obB) {
				if debugTraceGen {
					fmt.Printf("\tassuming %s as prime(%s) from obB\n", g.primer.Prime(m), m)
				}
				g.sat.Assume(g.primer.Prime(m))
			}
		}
		res := g.sat.Try()
		switch res {
		case 0:
			return trace, fmt.Errorf("ErrTraceBuildTimeout")
		case -1:
			return nil, fmt.Errorf("ErrInternalBadTrace: %v", g.sat.Why(nil))
		}
		for i := range vals {
			m := z.Var(i).Pos()
			vals[i] = g.sat.Value(m)
		}
		trace.Append(vals[:g.orgLen])
		obA = obB
		if obA == 0 {
			break
		}
		obB = g.obs.Parent(obB)
		for i, m := range g.trans.Latches {
			mp := g.primer.Prime(m)
			if !g.sat.Value(mp) {
				m = m.Not()
			}
			if debugTraceGen {
				fmt.Printf("for next step, assuming %s since prime(%s)=%s=%t\n", m,
					g.trans.Latches[i], mp, g.sat.Value(mp))
			}
			g.sat.Assume(m)
		}
	}
	return trace, nil
}
