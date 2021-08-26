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

package lits

import (
	"testing"

	"github.com/go-air/gini/z"
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
