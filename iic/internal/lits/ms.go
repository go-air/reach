package lits

import (
	"github.com/irifrance/gini/z"
)

func CalcSig(ms []z.Lit) (sig uint64) {
	for _, m := range ms {
		sig |= 1 << (uint64(m.Var()) % 64)
	}
	return sig
}

func ContainedBySorted(ms, ns []z.Lit) bool {
	i, j := 0, 0
	msz, nsz := len(ms), len(ns)
	var m, n z.Lit
	for i < msz && j < nsz {
		m = ms[i]
		n = ns[j]
		j++
		if m == n {
			i++
		}
	}
	return i == msz
}

func ContainedBySortedExcept(ms, ns []z.Lit, except z.Lit) bool {
	i, j := 0, 0
	msz, nsz := len(ms), len(ns)
	var m, n z.Lit
	for i < msz && j < nsz {
		m = ms[i]
		if m == except {
			i++
			continue
		}
		n = ns[j]
		j++
		if m == n {
			i++
		}
	}
	return i == msz
}

func Flip(ms []z.Lit) []z.Lit {
	for i, m := range ms {
		ms[i] = m.Not()
	}
	return ms
}

func MinLit(ms []z.Lit, measure func(m z.Lit) int) z.Lit {
	minM := ms[0]
	min := measure(minM)
	N := len(ms)
	var m z.Lit
	var d int
	for i := 1; i < N; i++ {
		m = ms[i]
		d = measure(m)
		if d < min {
			min = d
			minM = m
		}
	}
	return minM
}
