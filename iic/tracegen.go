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

package iic

import (
	"fmt"

	"github.com/go-air/gini/logic"
	"github.com/go-air/gini/z"
	"github.com/go-air/reach"
	"github.com/go-air/reach/iic/internal/obs"
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
