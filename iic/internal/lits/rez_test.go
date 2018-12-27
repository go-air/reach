package lits

import (
	"testing"

	"github.com/irifrance/gini/z"
)

func TestResolver(t *testing.T) {
	r := &Resolver{}
	var ms = []z.Lit{z.Var(1).Pos(), z.Var(2).Pos(), z.Var(3).Pos()}

	if r.Set(ms, z.Var(4)) {
		t.Errorf("accepted wrong pivot\n")
	}

	if !r.Set(ms, z.Var(1)) {
		t.Errorf("didn't accept correct pivot\n")
	}

	out, ok := r.Resolve(nil, []z.Lit{z.Var(1).Neg(), z.Var(4).Neg()})
	if !ok {
		t.Fatalf("didn't resolve")
	}
	if len(out) != 3 {
		t.Errorf("wrong result: %v\n", out)
	}

	out, ok = r.Resolve(out[:0], []z.Lit{z.Var(1).Neg(), z.Var(2).Neg()})
	if ok {
		t.Fatalf("ok'd resolution of tautology\n")
	}

	out, ok = r.Resolve(out[:0], []z.Lit{z.Var(2).Pos(), z.Var(1).Neg()})
	if !ok {
		t.Errorf("resolution not ok 2")
	}
	if len(out) != 2 {
		t.Errorf("wrong number of lits in output: %v\n", out)
	}
}
