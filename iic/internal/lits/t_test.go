package lits_test

import (
	"math/rand"
	"testing"

	"github.com/irifrance/gini/z"
	"github.com/irifrance/reach/iic/internal/lits"
)

func TestT(t *testing.T) {
	d := lits.New()
	N := 16384
	P := 0.51
	V := 2048
	L := 23
	O := 1
	ms := make([]z.Lit, L+O)
	sps := make([]lits.Span, 0, 1024)
	R := make(map[lits.Span][]z.Lit, 1024)
	sz := 0
	for i := 0; i < N; i++ {
		if i%100 == 0 {
			t.Logf("P %f\n", P)
		}
		if rand.Float64() <= P || len(sps) == 0 {
			n := rand.Intn(L) + O
			ms := ms[:n]
			for i := range ms {
				v := rand.Intn(V) + 1
				v *= 2
				if rand.Intn(2) == 1 {
					v += 1
				}
				ms[i] = z.Lit(v)
			}
			sp := d.Put(ms)
			sps = append(sps, sp)
			cs := make([]z.Lit, len(ms))
			copy(cs, ms)
			R[sp] = cs
			sz++
			continue
		}
		S := len(sps)
		if S != d.N() {
			t.Fatalf("op %d:  N %d != %d", i, S, d.N())
		}
		si := rand.Intn(S)
		span := sps[si]
		sps[si] = sps[S-1]
		sps = sps[:S-1]
		cs, ok := R[span]
		if !ok {
			t.Fatal("not in map")
		}
		ds := d.Get(span)
		if len(cs) != len(ds) {
			t.Errorf("op %d: len %d != %d |sps|=%d", i, len(cs), len(ds), len(sps))
			continue
		}
		for i := range ds {
			if ds[i] != cs[i] {
				t.Errorf("op %d: inequal s: %v %v", i, ds, cs)
				break
			}
		}
		d.Remove(span)
		if rand.Float64() >= 0.5 {
			delta := rand.Float64() - 0.5
			delta *= 0.01
			if P+delta > 0 && P+delta < 1.0 {
				P += delta
			}
		}
	}
}

func TestRepeatCompact(t *testing.T) {
	d := lits.New()
	N := 10
	sps := make([]lits.Span, 127)
	b := make([]z.Lit, 1)
	for n := 0; n < N; n++ {
		for i := range sps {
			b[0] = z.Var(i + 1).Pos()
			sps[i] = d.Put(b)
		}
		for i := range sps {
			if i%3 == 0 {
				continue
			}
			d.Remove(sps[i])
		}
		for i := range sps {
			if i%3 != 0 {
				continue
			}
			sp := sps[i]
			m := z.Var(i + 1).Pos()
			if d.Get(sp)[0] != m {
				t.Fatalf("mismatch")
			}
			d.Remove(sp)
		}
		if len(d.Get(0)) != 0 {
			t.Errorf("zero has %v\n", d.Get(0))
		}
	}
}
