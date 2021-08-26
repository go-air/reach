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

package obs

import (
	"fmt"
	"math/rand"
	"sort"
	"testing"

	"github.com/go-air/gini/gen"
	"github.com/go-air/gini/z"
	"github.com/go-air/reach/iic/internal/lits"
)

func TestSetFilt(t *testing.T) {
	D := lits.New()
	set := NewSet(D)
	NVars := 16
	KStar := 8
	for set.MaxK() < KStar {
		set.Grow()
	}
	M := 20

	cuber := gen.NewRandCuber(NVars/2, z.Var(NVars+1))
	var ms []z.Lit
	var bs []z.Lit

	k := KStar
	m := 0
	for k > 0 {
		id := set.Choose()
		fmt.Printf("chose %s\n", set.String(id))
		if id == 0 {
			t.Fatal("got null")
			continue
		}
		if set.IsRoot(id) || set.K(id) >= k {
			ms = cuber.RandCube(ms[:0])
			set.Extend(id, ms, ms[0])
			continue
		}
		bs = block(bs, set.Ms(id))
		fmt.Printf("block %s with %v\n", set.String(id), bs)
		set.Block(id, bs)
		m++
		if m == M-1 {
			k--
			m = 0
		}
	}

	for {
		id := set.Choose()
		fmt.Printf("chose block to %s\n", set.String(id))
		if id == 0 {
			t.Fatal("got null")
			continue
		}
		if set.IsRoot(id) {
			break
		}
		bs = block(bs, set.Ms(id))
		fmt.Printf("block %s with %v\n", set.String(id), bs)
		set.Block(id, bs)
	}
	if set.Choose() != 0 {
		t.Errorf("didn't finish on 0.\n")
	}
}

func TestNext(t *testing.T) {
	D := lits.New()
	set := NewSet(D)
	NVars := 387
	cuber := gen.NewRandCuber(NVars/2, z.Var(NVars+1))
	var ms, bs []z.Lit
	N := 1024
	for n := 0; n < N; n++ {
		set.Grow()
	}
	for n := 0; n < N; n++ {
		id := set.Choose()
		ms = cuber.RandCube(ms[:0])
		set.Extend(id, ms, ms[0])
	}
	for n := 0; n < N; n++ {
		id := set.Choose()
		bs = block(bs, set.Ms(id))
		set.Block(id, bs)
	}
	set.Grow()
	for n := 0; n < N; n++ {
		id := set.Choose()
		if id == 0 {
			t.Error("premature 0")
			break
		}
		if set.IsRoot(id) {
			t.Error("premature root")
			break
		}
	}
}

func block(dst []z.Lit, ms []z.Lit) []z.Lit {
	if len(ms) == 0 {
		panic("wilma!")
	}
	if len(ms) == 1 {
		return []z.Lit{ms[0]}
	}
	n := rand.Intn(len(ms)) + 1
	if cap(dst) < n {
		dst = make([]z.Lit, 0, n*2)
	}
	dst = dst[:0]
	done := make(map[z.Lit]bool, n)
	if n == len(ms) {
		dst = append(dst, ms...)
		goto sortIt
	}
	for len(dst) < n {
		c := rand.Intn(len(ms))
		m := ms[c]
		if _, inDone := done[m]; inDone {
			continue
		}
		done[m] = true
		dst = append(dst, m)
	}
sortIt:
	sort.Slice(dst, func(i, j int) bool {
		return dst[i] < dst[j]
	})
	return lits.Flip(dst)
}
