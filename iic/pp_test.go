package iic

import (
	"fmt"
	"os"
	"testing"

	"github.com/irifrance/gini/logic"
	"github.com/irifrance/gini/z"
)

func TestPpElim(t *testing.T) {
	trans := logic.NewS()
	m := trans.Latch(trans.F)
	a, b, c := trans.Lit(), trans.Lit(), trans.Lit()
	o := trans.Ands(a, b, c)
	trans.SetNext(m, o)
	p := newPp(trans, m.Not())
	trans.ToCnf(p)
	p.dump(os.Stdout)
	dc, dm := p.tryElim(a)
	t.Logf("%s: %d, %d", a, dc, dm)
	p.elim(a)
	p.dump(os.Stdout)
}

func TestPpSubsume(t *testing.T) {
	trans := logic.NewS()
	for i := 0; i < 16; i++ {
		trans.Lit()
	}
	p := newPp(trans, z.Var(1).Pos())
	orgNumClauses := p.numClauses()

	p.Add(z.Var(7).Pos())
	p.Add(0)
	p.Add(z.Var(7).Pos())
	p.Add(z.Var(11).Pos())
	p.Add(0)

	p.ssr()
	if p.numClauses()-orgNumClauses != 1 {
		t.Errorf("didn't subsume: orgclauses %d added 2, after subsume have %d\n", orgNumClauses, p.numClauses())
	}
}

func TestPpSsr(t *testing.T) {
	trans := logic.NewS()
	for i := 0; i < 16; i++ {
		trans.Lit()
	}
	p := newPp(trans, z.Var(1).Pos())

	p.Add(z.Var(2).Pos())
	p.Add(z.Var(3).Pos())
	p.Add(z.Var(4).Pos())
	p.Add(0)
	p.Add(z.Var(2).Neg())
	p.Add(z.Var(3).Pos())
	p.Add(z.Var(4).Pos())
	p.Add(z.Var(5).Pos())
	p.Add(0)
	p.Add(z.Var(2).Neg())
	p.Add(z.Var(3).Neg())
	p.Add(z.Var(5).Pos())
	p.Add(0)
	p.Add(z.Var(3).Pos())
	p.Add(z.Var(4).Pos())
	p.Add(z.Var(5).Pos())
	p.Add(z.Var(6).Pos())
	p.Add(0)

	fmt.Printf("b4 ssr:\n")
	p.dump(os.Stdout)
	fmt.Printf("after ssr:\n")

	orgNumClauses := p.numClauses()
	p.ssr()
	if orgNumClauses-p.numClauses() != 1 {
		t.Errorf("ssr?")
	}
	p.dump(os.Stdout)
}
