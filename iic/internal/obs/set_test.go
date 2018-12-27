package obs

import (
	"fmt"
	"math/rand"
	"sort"
	"testing"

	"github.com/irifrance/gini/gen"
	"github.com/irifrance/gini/z"
	"github.com/irifrance/reach/iic/internal/lits"
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
