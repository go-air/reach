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
	"fmt"

	"github.com/go-air/gini/z"
)

type Resolver struct {
	values []int8
	ms     []z.Lit
	pivot  z.Var
}

func NewResolver(capHint int) *Resolver {
	return &Resolver{
		values: make([]int8, capHint),
		ms:     make([]z.Lit, 0, capHint)}
}

func (r *Resolver) Set(ms []z.Lit, pivot z.Var) bool {
	for _, m := range r.ms {
		r.values[m.Var()] = 0
	}
	for _, m := range ms {
		r.ensureM(m)
		r.values[m.Var()] = m.Sign()
	}
	r.ms = r.ms[:0]
	r.ms = append(r.ms, ms...)
	return r.SetPivot(pivot)
}

func (r *Resolver) SetPivot(pivot z.Var) bool {
	for i, m := range r.ms {
		if m.Var() == pivot {
			r.pivot = pivot
			r.ms[0], r.ms[i] = r.ms[i], r.ms[0]
			return true
		}
	}
	return false
}

func (r *Resolver) Resolve(dst []z.Lit, os []z.Lit) (out []z.Lit, ok bool) {
	orgLen := len(dst)
	var mv z.Var
	var osPivot z.Var
	for _, m := range os {
		r.ensureM(m)
		mv = m.Var()

		if r.values[mv]+m.Sign() == 0 {
			if mv == r.pivot {
				osPivot = mv
				continue
			}
			dst = dst[:orgLen]
			return dst, false
		}
		if r.values[mv] == 0 {
			dst = append(dst, m)
		}
	}
	if osPivot == 0 {
		dst = dst[:orgLen]
		return dst, false
	}
	dst = append(dst, r.ms[1:]...)
	return dst, true
}

func (r *Resolver) ensureM(m z.Lit) {
	im := int(m)
	if im < len(r.values) {
		return
	}
	tmp := make([]int8, im*2)
	copy(tmp, r.values)
	r.values = tmp
}

func (r *Resolver) String() string {
	return fmt.Sprintf("resolver[pivot=%s ms=%v]", r.pivot.Pos(), r.ms)
}
