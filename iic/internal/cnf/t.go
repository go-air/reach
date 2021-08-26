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

package cnf

import (
	"fmt"
	"io"
	"sort"

	"github.com/go-air/gini"
	"github.com/go-air/gini/inter"
	"github.com/go-air/gini/z"
	"github.com/go-air/reach/iic/internal/lits"
)

const (
	// causes sanity checks that shouldn't be necessary
	sanitize = true
)

// T represents a multi-level CNF for iic.
type T struct {
	clauses []clause // holds clauses indexed by Id.
	ids     [][]Id   // ids[k] at level k, dirty and cleaned up lazily
	occs    [][]Id   // shared, built one at a time per level
	free    []Id
	sat     *gini.Gini // adding/activation coordination
	lits    *lits.T    // storage of literals.
	rmHook  func(f *T, rm, by Id, K int)
}

// New creates a new IIC multi-level CNF from a
// lit store and a sat solver.
func New(sat *gini.Gini, d *lits.T) *T {
	return &T{sat: sat, lits: d, clauses: make([]clause, 1, 1024)}
}

// Add adds the clause representing the disjunction of
// ms at level k.
//
// Add panics if k >= f.K().
func (f *T) Add(ms []z.Lit, k int) Id {
	sort.Slice(ms, func(i, j int) bool {
		return ms[i] < ms[j]
	})
	if sanitize {
		for i := 1; i < len(ms); i++ {
			if ms[i] == ms[i-1] {
				panic("ErrSanitize")
			}
		}
	}
	cls := f.newClause()
	f.ids[k] = append(f.ids[k], cls.id)
	cls.ms = f.lits.Put(ms)
	cls.sig = lits.CalcSig(ms)
	cls.level = k
	cls.rm = false
	sat := f.sat
	for _, m := range ms {
		sat.Add(m)
	}
	cls.act = sat.Activate()
	return cls.id
}

func (f *T) String(c Id) string {
	return fmt.Sprintf("%s %v", c, f.Lits(c))
}

// Lits returns the clauses' literals, which
// are backed by a unified literal store for iic.
func (f *T) Lits(c Id) []z.Lit {
	cl := &f.clauses[c]
	return f.lits.Get(cl.ms)
}

// Len returns the length of c.Lits()
func (f *T) Len(c Id) int {
	cl := &f.clauses[c]
	return f.lits.Len(cl.ms)
}

func (f *T) Level(c Id) int {
	cls := &f.clauses[c]
	return cls.level
}

func (f *T) ActLit(c Id) z.Lit {
	cls := &f.clauses[c]
	return cls.act
}

// K gives the max level present in the cnf `f`.
func (f *T) K() int {
	return len(f.ids) - 1
}

// PushK extends the CNF one level.
func (f *T) PushK() {
	f.ids = append(f.ids, []Id{})
}

// Len returns the number of clauses in `f`.
func (f *T) NumClauses() int {
	return len(f.clauses) - len(f.free) - 1
}

// LenK returns the size of level k.
func (f *T) LenK(k int) int {
	ttl := 0
	f.Forall(k, func(_ *T, _ Id) {
		ttl++
	})
	return ttl
}

// SetRemoveHook allows to specify a function callback
// when a clause is removed by simplfication.
func (f *T) SetRemoveHook(fn func(f *T, rm, by Id, K int)) {
	f.rmHook = fn
}

// Forall iterates over all clauses at level k.
func (f *T) Forall(k int, fn func(f *T, cls Id)) {
	j := 0
	var cls *clause
	ids := f.ids[k]
	for _, id := range ids {
		cls = &f.clauses[id]
		if cls.rmd() || cls.level != k {
			continue
		}
		fn(f, id)
		ids[j] = id
		j++
	}
	f.ids[k] = ids[:j]
}

// Push pushes the clause `c` forward one level.
// If `c` is at the max level `f.K()`, then
// `c` is queued for the next `PushK`
func (f *T) Push(c Id) {
	cls := &f.clauses[c]
	cls.level++
	if cls.level >= len(f.ids) {
		f.ids = append(f.ids, []Id{})
	}
	f.ids[cls.level] = append(f.ids[cls.level], c)
}

// AssumeLevel causes the SAT solver in New
// to assume all clauses at level k.
func (f *T) AssumeLevel(k int) {
	for i := k; i < len(f.ids); i++ {
		f.assumeInternalLevel(i)
	}
}

// Simplify simplifies the clauses at level `k`,
// removing redundant ones.
func (f *T) Simplify(k int) {
	occs := f.kOccs(k)
	ids := f.ids[k]
	var cid Id
	var cls *clause
	for _, cid = range ids {
		cls = &f.clauses[cid]
		if cls.rmd() || cls.level != k {
			continue
		}
		f.removeSubsumed(cls, occs, false)
	}
	j := 0
	for _, cid = range ids {
		cls = &f.clauses[cid]
		if cls.rmd() || cls.level != k {
			continue
		}
		ids[j] = cid
		j++
	}
	f.ids[k] = ids[:j]
}

func (f *T) RemoveDups(k int) {
	occs := f.kOccs(k)
	ids := f.ids[k]
	var cid Id
	var cls *clause
	for _, cid = range ids {
		cls = &f.clauses[cid]
		if cls.rmd() || cls.level != k {
			continue
		}
		f.removeSubsumed(cls, occs, true)
	}
	j := 0
	for _, cid = range ids {
		cls = &f.clauses[cid]
		if cls.rmd() || cls.level != k {
			continue
		}
		ids[j] = cid
		j++
	}
	f.ids[k] = ids[:j]
}

func (f *T) kOccs(k int) [][]Id {
	occs := f.occs
	for i := range occs {
		occs[i] = occs[i][:0]
	}
	for _, n := range f.ids[k] {
		c := &f.clauses[n]
		if c.rmd() || c.level != k {
			continue
		}
		for _, m := range f.Lits(n) {
			for int(m) >= len(occs) {
				occs = append(occs, []Id{})
			}
			occs[m] = append(occs[m], n)
		}
	}
	return occs
}

func (f *T) Stats(dst io.Writer) {
	N := len(f.ids)
	i := N - 20
	if i < 0 {
		i = 0
	}
	ncs := make([]int, 0, 7)
	nms := make([]int, 0, 7)
	j := i
	for i < N {
		nc := 0
		nm := 0
		f.Forall(i, func(f *T, c Id) {
			nc++
			nm += f.Len(c)
		})
		ncs = append(ncs, nc)
		nms = append(nms, nm)
		i++
	}
	fmt.Fprintf(dst, "cnf %d clauses %d levels [%d..%d):\n\t", f.NumClauses(), N, j, N)
	for _, nc := range ncs {
		fmt.Fprintf(dst, "%05d ", nc)
	}
	fmt.Fprintf(dst, "\n\t")
	for _, nm := range nms {
		fmt.Fprintf(dst, "%05d ", nm)
	}
	fmt.Fprintf(dst, "\n")
}

// DimacsHdrInfo calculates the header info
// for iic level k.
func (f *T) DimacsHdrInfoAt(k int) (maxVar, nClauses int) {
	v, nc := 0, 0
	N := len(f.ids)
	for j := k; j < N; j++ {
		f.Forall(k, func(f *T, c Id) {
			nc++
			for _, m := range f.Lits(c) {
				mv := int(m.Var())
				if mv > v {
					v = mv
				}
			}
		})
	}
	return v, nc
}

func (f *T) DimacsHdrInfo() (maxVar, nClauses int) {
	v, nc := 0, 0
	for i := 0; i < len(f.ids); i++ {
		kv, knc := f.DimacsHdrInfoAt(i)
		if kv > v {
			v = kv
		}
		nc += knc
	}
	return v, nc
}

// WriteLevel writes the cnf level k
func (f *T) WriteLevel(dst io.Writer, k int, prefix string) (int, error) {
	var ttl, n int
	var err error
	f.Forall(k, func(f *T, c Id) {
		if err != nil {
			return
		}
		ms := f.Lits(c)
		N := len(ms)
		n, err = dst.Write([]byte(prefix))
		ttl += n
		if err != nil {
			return
		}
		/*

			n, err = dst.Write([]byte(fmt.Sprintf("c %s\n%s", c, prefix)))
			ttl += n
			if err != nil {
				return
			}
		*/

		var s string
		for i, m := range ms {
			if i == N-1 {
				s = fmt.Sprintf("%s", m)
			} else {
				s = fmt.Sprintf("%s ", m)
			}
			n, err = dst.Write([]byte(s))
			ttl += n
			if err != nil {
				return
			}
		}
		n, err = dst.Write([]byte("\n"))
		ttl += n
	})
	return ttl, err
}

type adder struct {
	k  int
	f  *T
	ms []z.Lit
}

func (a *adder) Add(m z.Lit) {
	if m == z.LitNull {
		a.f.Add(a.ms, a.k)
		a.ms = a.ms[:0]
		return
	}
	a.ms = append(a.ms, m)
}

func (f *T) Adder(k int) inter.Adder {
	return &adder{k: k, f: f}
}

// Dump dumps `f` to dst, indicating each
// level.
func (f *T) Dump(dst io.Writer) (int, error) {
	V, C := f.DimacsHdrInfo()
	var n, ttl int
	var err error
	var s string
	s = fmt.Sprintf("p cnf %d %d\n", V, C)
	n, err = dst.Write([]byte(s))
	ttl += n
	if err != nil {
		return ttl, err
	}
	N := len(f.ids)
	for k := 0; k < N; k++ {
		n, err = fmt.Fprintf(dst, "c level %d:\n", k)
		ttl += n
		if err != nil {
			return ttl, err
		}
		n, err = f.WriteLevel(dst, k, "")
		ttl += n
		if err != nil {
			return ttl, err
		}
	}
	return ttl, err
}

func (f *T) assumeInternalLevel(k int) {
	j := 0
	var cls *clause
	ids := f.ids[k]
	for _, id := range ids {
		cls = &f.clauses[id]
		if cls.rmd() || cls.level != k {
			continue
		}
		f.sat.Assume(cls.act)
		ids[j] = id
		j++
	}
	f.ids[k] = ids[:j]
}

func (f *T) removeSubsumed(cls *clause, occs [][]Id, dupsOnly bool) {
	ms := f.Lits(cls.id)
	k := cls.level

	j := 0
	var ocls *clause
	minM := lits.MinLit(ms, func(m z.Lit) int {
		return len(occs[m])
	})
	mOccs := occs[minM]
	var oms []z.Lit

	for _, n := range mOccs {
		ocls = &f.clauses[n]
		if ocls.level != k {
			continue
		}
		if ocls.rmd() {
			continue
		}
		if ocls == cls {
			mOccs[j] = n
			j++
			continue
		}
		if (ocls.sig & cls.sig) != cls.sig {
			mOccs[j] = n
			j++
			continue
		}
		oms = f.Lits(n)
		if len(oms) < len(ms) {
			mOccs[j] = n
			j++
			continue
		}
		if dupsOnly && len(oms) > len(ms) {
			mOccs[j] = n
			j++
			continue
		}
		if !lits.ContainedBySorted(ms, oms) {
			mOccs[j] = n
			j++
			continue
		}
		if f.rmHook != nil {
			f.rmHook(f, ocls.id, cls.id, f.K())
		}
		f.sat.Deactivate(ocls.act)
		f.setRm(n)
		f.free = append(f.free, n)
	}
	occs[minM] = mOccs[:j]
}

func (f *T) newClause() *clause {
	cid := f.cid()
	c := &f.clauses[cid]
	c.id = cid
	return c
}

func (f *T) cid() Id {
	N := len(f.free)
	if N > 0 {
		N--
		id := f.free[N]
		f.free = f.free[:N]
		return Id(id)
	}
	N = len(f.clauses)
	if N == cap(f.clauses) {
		tmp := make([]clause, N*2)
		copy(tmp, f.clauses)
		f.clauses = tmp
	}
	f.clauses = f.clauses[:N+1]
	return Id(N)
}

func (f *T) setRm(c Id) {
	cls := &f.clauses[c]
	cls.rm = true
	f.lits.Remove(cls.ms)
	cls.ms = 0
}
