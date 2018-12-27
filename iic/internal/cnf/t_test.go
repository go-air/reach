package cnf_test

import (
	"fmt"
	"testing"

	"github.com/irifrance/gini/z"
	"github.com/irifrance/reach/iic/internal/lits"

	"github.com/irifrance/gini"
	"github.com/irifrance/reach/iic/internal/cnf"
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
