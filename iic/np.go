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
	"io"
	"math"

	"github.com/go-air/gini/z"
	"github.com/go-air/reach"
	"github.com/go-air/reach/iic/internal/cnf"
	"github.com/go-air/reach/iic/internal/obs"
)

type np struct {
	cnf      *cnf.T
	sat      *satmon
	pri      *reach.Primer
	obs      *obs.Set
	initVals []int8
	levels   []pl
	init     z.Lit
	bad      z.Lit
	k        int

	conSift     bool
	conSiftPull bool

	nBlock  int64
	nExtend int64
}

func newNp(cnf *cnf.T, sat *satmon, pri *reach.Primer, obs *obs.Set, initVals []int8, init, bad z.Lit) *np {
	res := &np{}
	res.sat = sat
	res.cnf = cnf
	res.pri = pri
	res.obs = obs
	res.init = init
	res.bad = bad
	res.levels = make([]pl, 0, 1024)
	res.initVals = initVals
	return res
}

func (p *np) onBlock(c cnf.Id) {
	p.nBlock++
	k := p.cnf.Level(c)
	lvl := p.level(k)
	lvl.nBlock++
	if !configConSift && !configConSiftPull {
		return
	}

	cnfLen := p.cnf.LenK(k)
	potential := cnfLen - lvl.lastSiftLen
	if false && p.cnf.K() == k && potential < 100 {
		return
	}
	alpha := 0.95
	cmp := float64(cnfLen) * math.Pow(alpha, float64(len(p.levels)-k))
	icmp := int(math.Floor(cmp + 0.5))

	if potential*3 <= icmp {
		return
	}
	if p.conSift {
		lvl.conSift(p.cnf, p.sat, p.obs, p.init, p.bad)
		lvl.lastSiftLen = p.cnf.LenK(k)
		lvl.lastnBlock = p.nBlock
	}
	if p.conSiftPull && k > 1 {
		lvl := p.level(k - 1)
		lvl.prop(p.cnf, p.sat, p.bad, p.pri, p.obs)
	}
}

func (p *np) onExtend(k int) {
	p.nExtend++
	p.levels[k].nExtend++
}

func (p *np) prop(k int) (fixedPoint, timeOk bool) {
	if debugPushes {
		fmt.Printf("prop %d len %d\n", k, p.cnf.LenK(k-1))
	}
	for j := p.k; j < k; j++ {
		lvl := p.level(j)
		if !lvl.prop(p.cnf, p.sat, p.bad, p.pri, p.obs) {
			return false, false
		}
	}
	if configCnfSimplify {
		p.cnf.Simplify(k)
	}
	if debugPushes {
		fmt.Printf("after prop %d len %d\n", k, p.cnf.LenK(k-1))
	}
	p.k = k
	return p.cnf.LenK(k-1) == 0, true
}

func (p *np) Stats(w io.Writer) {
	fmt.Fprintf(w, "pushes:\n")
	for i := range p.levels {
		lvl := p.level(i)
		fmt.Fprintf(w, "\tlevel %04d: %04d sifts %04d clauses %04d reduced\n", i,
			lvl.sifts, lvl.siftAttempts, lvl.siftReduced)
	}
}

func (p *np) push() {
	n := len(p.levels)
	p.levels = append(p.levels, pl{})
	lvl := p.level(n)
	lvl.k = n
	lvl.sifter = newSifter(p.cnf, p.sat, p.pri, p.initVals)
	p.k = 0
	for i := range p.levels {
		lvl := p.level(i)
		lvl.lastSiftLen = p.cnf.LenK(i)
	}
}

func (p *np) crmHook(f *cnf.T, c, by cnf.Id, kstar int) {
	k := f.Level(c)
	if k >= len(p.levels) {
		return
	}
	p.levels[k].crmHook(c)
}

func (p *np) level(k int) *pl {
	return &p.levels[k]
}
