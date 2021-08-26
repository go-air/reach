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
	"github.com/go-air/gini/z"
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
