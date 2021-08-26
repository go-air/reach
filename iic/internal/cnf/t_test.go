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

package cnf_test

import (
	"fmt"
	"testing"

	"github.com/go-air/gini/z"
	"github.com/go-air/reach/iic/internal/lits"

	"github.com/go-air/gini"
	"github.com/go-air/reach/iic/internal/cnf"
)

func TestT(t *testing.T) {
	s := gini.New()
	d := lits.New()
	f := cnf.New(s, d)
	f.PushK()
	if f.K() != 0 {
		t.Errorf("K %d", f.K())
	}
	c := f.Add([]z.Lit{4, 6, 9}, 0)
	if f.Len(c) != 3 {
		t.Errorf("AddLen %d", f.Len(c))
	}
	b := f.Add([]z.Lit{4, 6}, 0)
	if f.Len(b) != 2 {
		t.Errorf("AddLen2 %d", f.Len(b))
	}
	fmt.Printf("c: %s\nb: %s\n", c, b)
	rmd := false
	f.SetRemoveHook(func(f *cnf.T, cls, _ cnf.Id, _ int) {
		rmd = true
		if cls != c {
			t.Errorf("rm got %s not %s", cls, b)
		}
	})
	f.Simplify(0)
	if !rmd {
		t.Errorf("subsumed not rmd")
	}
	if f.NumClauses() != 1 {
		t.Errorf("len after subsume: %d", f.NumClauses())
	}
	f.PushK()
	f.Push(b)
	if f.Level(b) != 1 {
		t.Errorf("after push got %d", f.Level(b))
	}
	f.Forall(0, func(_ *cnf.T, c cnf.Id) {
		t.Errorf("something in 0")
	})
	f.Forall(1, func(_ *cnf.T, c cnf.Id) {
		if c != b {
			t.Errorf("got %s not %s", c, b)
		}
	})
	c = f.Add([]z.Lit{4, 6, 9}, 0)
	f.SetRemoveHook(func(f *cnf.T, c, _ cnf.Id, _ int) {
		t.Errorf("removing %s", c)
	})
	f.Simplify(0)
}

func TestSimplifyEq(t *testing.T) {
	s := gini.New()
	d := lits.New()
	f := cnf.New(s, d)
	f.PushK()
	f.PushK()
	m, n := s.Lit(), s.Lit()
	f.Add([]z.Lit{m, n}, 1)
	f.Add([]z.Lit{m, n}, 1)
	nRm := 0
	f.SetRemoveHook(func(f *cnf.T, c, by cnf.Id, maxK int) {
		nRm++
	})
	f.Simplify(1)
	if f.LenK(1) != 1 {
		t.Errorf("simplify didn't work.")
	}
	if nRm != 1 {
		t.Errorf("simplify didn't call hook\n")
	}
}

func TestAssumeLevel(t *testing.T) {
	s := gini.New()
	d := lits.New()
	f := cnf.New(s, d)
	f.PushK()
	f.PushK()
	f.PushK()
	m, n := s.Lit(), s.Lit()
	f.Add([]z.Lit{m}, 1)
	f.Add([]z.Lit{m.Not(), n}, 1)
	f.Add([]z.Lit{n.Not()}, 2)
	f.AssumeLevel(1)
	if s.Solve() == 1 {
		t.Errorf("cnf assume level sat")
	}
	f.AssumeLevel(2)
	if s.Solve() == -1 {
		t.Errorf("cnf assume level unsat")
	}
}
