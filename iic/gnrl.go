// Copyright 2018 The Reach Authors. All rights reserved.  Use of this source
// code is governed by a license that can be found in the License file.

package iic

import (
	"fmt"
	"io"
	"log"
	"math/rand"

	"github.com/go-air/gini/logic"
	"github.com/go-air/gini/z"
	"github.com/go-air/reach"
	"github.com/go-air/reach/iic/internal/lits"
	"github.com/go-air/reach/iic/internal/obs"
)

// handles generalization of inductive clauses.
type gnrl struct {
	lits   *lits.T
	obs    *obs.Set
	sat    *satmon
	minSet []int8
	whys   []z.Lit
	ns, os []z.Lit

	initVals []int8
	initM    z.Lit

	doRemoveLits bool

	nLitsIn  int64
	nSift    int64
	nC       int64
	nStep    int64
	nRm      int64
	nTryRm   int64
	nTryFree int64
}

func newGnrl(sat *satmon, trans *logic.S, d *lits.T, obs *obs.Set, initVals []int8) *gnrl {
	res := &gnrl{sat: sat, lits: d, obs: obs}
	res.minSet = make([]int8, trans.Len())
	res.initVals = initVals
	return res
}

// called under test scope for o with cnf at o.minLevel-1
// and not(o.ms) under activation at o.minLevel.
// above this scope, the primes of o.ms have been assumed,
// and result is unsat.
func (g *gnrl) gnrlize(o obs.Id, primer *reach.Primer) bool {
	defer func() {
		if res := g.sat.sat.Untest(); res == -1 {
			panic("untest handle ob unsat ind.")
		}
	}()
	g.nC++
	g.ns = g.ns[:0]
	g.os = g.os[:0]
	g.os = append(g.os, g.obs.Ms(o)...)
	orgLen := int64(len(g.os))
	g.initM = g.obs.InitWit(o)
start:
	g.nStep++
	if debugGnrl {
		fmt.Printf("gnrl: %s %v i%s\n", o, g.os, g.initM)
	}
	g.whys = g.whys[:0]
	g.whys = g.sat.Why(g.whys)

	for _, m := range g.whys {
		g.ensureMinSet(m)
		g.minSet[m.Var()] = m.Sign()
	}
	g.ns = g.ns[:0]
	var whyInit z.Lit
	for _, m := range g.os {
		mp := primer.Prime(m)
		mpv := mp.Var()
		if g.minSet[mpv] == 0 {
			continue
		}
		g.ns = append(g.ns, m)
		if whyInit != z.LitNull {
			continue
		}
		if g.initVals[m.Var()]+m.Sign() == 0 {
			whyInit = m
		}
	}
	if whyInit == z.LitNull {
		if debugGnrl || debugInits {
			fmt.Printf("for %v no whys satisfy init. ", g.obs.Ms(o))
		}
		if g.initM != z.LitNull {
			g.ns = append(g.ns, g.initM)
			if debugGnrl || debugInits {
				fmt.Printf("backing off to prime(o initWitness) %s -> %s\n",
					g.initM, primer.Prime(g.initM))
			}
			orgLen++
		} else {
			// counterexample found but at wrong level
			log.Fatalf("ErrInternalGnrlNoInit %s %v", g.obs.String(o), g.ns)
		}
	}
	g.cleanupStep()
	if len(g.ns) >= len(g.os) {
		if debugGnrl {
			fmt.Printf("nothing removed, returning\n")
		}
		g.nSift += orgLen - int64(len(g.ns))
		g.nLitsIn += orgLen
		if g.doRemoveLits {
			return g.removeLits(primer)
		}
		return true
	}
	rand.Shuffle(len(g.ns), func(i, j int) {
		g.ns[i], g.ns[j] = g.ns[j], g.ns[i]
	})
	for _, m := range g.ns {
		mp := primer.Prime(m)
		g.sat.Assume(mp)
	}
	g.os, g.ns = g.ns, g.os
	res := g.sat.Try()
	if res == -1 {
		goto start
	}
	return res != 0
}

func (g *gnrl) removeLits(primer *reach.Primer) (ok bool) {
	ok = true
	g.placeInit(g.ns)

	n := len(g.ns)
	g.nTryRm += int64(n)
	failures := 0
	for len(g.ns) > 1 && ok && failures <= n/3 {
		guess := rand.Intn(len(g.ns)-1) + 1
		for j, m := range g.ns {
			mp := primer.Prime(m)
			if j == guess {
				//g.sat.Assume(mp.Not())
				continue
			}
			g.sat.Assume(mp)
		}
		switch g.sat.Try() {
		case 0:
			ok = false
		case 1:
			failures++
		case -1:
			// if we do this we need to union they whys of both sides...
			//n := len(g.ns) - 1
			//g.ns[guess], g.ns[n] = g.ns[n], g.ns[guess]
			//g.ns = g.ns[:n]
			g.whys = g.whys[:0]
			g.whys = g.sat.Why(g.whys)

			for _, m := range g.whys {
				g.ensureMinSet(m)
				g.minSet[m.Var()] = m.Sign()
			}
			j := 0
			for i, m := range g.ns {
				if i == 0 {
					j++
					continue
				}
				if i == guess {
					continue
				}
				mp := primer.Prime(m)
				if g.minSet[mp.Var()] == 0 {
					continue
				}
				g.ns[j] = m
				j++
			}
			for _, m := range g.whys {
				g.minSet[m.Var()] = 0
			}
			g.nTryFree += int64(len(g.ns)-j) - 1
			g.ns = g.ns[:j]
		}
	}
	g.nRm += int64(n - len(g.ns))
	return ok
}

func (g *gnrl) placeInit(ms []z.Lit) {
	for i, m := range ms {
		if g.initVals[m.Var()]+m.Sign() == 0 {
			ms[0], ms[i] = m, ms[0]
			break
		}
	}
	m := ms[0]
	if g.initVals[m.Var()]+m.Sign() != 0 {
		panic("no init in placeInit")
	}
}

func (g *gnrl) cnfMs() (ms []z.Lit, init z.Lit) {
	// flip for cnf
	for i, m := range g.ns {
		g.ns[i] = m.Not()
	}
	if ini := g.clsInitOk(g.ns); ini != z.LitNull {
		return g.ns, ini
	}
	if g.initM != z.LitNull {
		if debugGnrl {
			fmt.Printf("gnrl learning %v + %s\n", g.ns, g.initM)
		}
		init := g.initM.Not()
		g.ns = append(g.ns, init)
		return g.ns, init
	}
	log.Fatalf("gnrl: init not ok: g.ns %v\n", g.ns)
	return g.ns, z.LitNull
}

func (g *gnrl) clsInitOk(ms []z.Lit) z.Lit {
	iVals := g.initVals
	for _, m := range ms {
		if iVals[m.Var()] == m.Sign() {
			return m
		}
	}
	return z.LitNull
}

func (g *gnrl) cleanupStep() {
	for _, m := range g.whys {
		g.minSet[m.Var()] = 0
	}
}

func (g *gnrl) Stats(dst io.Writer) {
	fmt.Fprintf(dst, "gnrl: %d-c/%d-m  %d-ss/%d-sift %d-trym %d-rm/%d-rmy\n", g.nC, g.nLitsIn, g.nStep, g.nSift, g.nTryRm, g.nRm, g.nTryFree)
	g.sat.Stats(dst)
}

func (g *gnrl) ensureMinSet(m z.Lit) {
	n := int(m.Var()) + 1
	if n >= cap(g.minSet) {
		tmp := make([]int8, len(g.minSet), n*5/3)
		copy(tmp, g.minSet)
		g.minSet = tmp
	}
	if n >= len(g.minSet) {
		g.minSet = g.minSet[:n]
	}
}
