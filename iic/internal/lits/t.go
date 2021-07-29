package lits

import (
	"fmt"
	"sort"

	"github.com/go-air/gini/z"
)

// T represents a contiguous store of literals.
type T struct {
	D     []z.Lit
	spans []mspan
	rmq   []Span
	free  []Span
	nFree int // lits
}

type mspan struct {
	start, end int
}

func (m *mspan) String() string {
	return fmt.Sprintf("[%d..%d)", m.start, m.end)
}

// Span represents a span of literals in T.
type Span int

func (s *mspan) size() int {
	return s.end - s.start
}

// New creates a new store of lits.
func New() *T {
	d := &T{D: make([]z.Lit, 1, 768), spans: make([]mspan, 1, 128)}
	return d
}

// Put copies `ms` into `d` and returns the associated Span.
func (d *T) Put(ms []z.Lit) Span {
	start := len(d.D)
	d.D = append(d.D, ms...)
	end := len(d.D)
	sp, n := d.mspan()
	sp.start = start
	sp.end = end
	return n
}

// Get returns a slice backed by the contiguous storage
// in `d` containing the lits represented by `n`.
func (d *T) Get(n Span) []z.Lit {
	s := &d.spans[n]
	return d.D[s.start:s.end]
}

// Remove removes the span `n`
// `n` must no longer be referenced by the user of `d`.
func (d *T) Remove(n Span) {
	if n == 0 {
		panic("ErrInternalLitsRemoveNil")
	}
	sp := &d.spans[n]
	d.nFree += sp.size()
	d.rmq = append(d.rmq, n)
	if 2*d.nFree < len(d.D) {
		return
	}
	d.compact()
}

// Len returns the number of literals in `n`
func (d *T) Len(n Span) int {
	sp := &d.spans[n]
	return sp.end - sp.start
}

// N returns the number of spans in `d`.
func (d *T) N() int {
	return len(d.spans) - len(d.rmq) - len(d.free) - 1
}

func (d *T) compact() {
	// nb pre condition is that d.nFree != 0
	sps := make([]Span, len(d.spans)-1)
	for i := range sps {
		sps[i] = Span(i + 1)
	}
	sort.Slice(sps, func(i, j int) bool {
		si, sj := sps[i], sps[j]
		spi, spj := &d.spans[si], &d.spans[sj]
		if spi.start < spj.start {
			return true
		}
		if spi.start == spj.start {
			return false
		}
		if spi.end < spj.end {
			return true
		}
		if spi.end > spj.end {
			return false
		}
		return si < sj
	})
	var msp *mspan
	for _, spi := range d.rmq {
		msp = &d.spans[spi]
		msp.start = -msp.start
	}
	top := 1
	var start, end, sz int

	D := d.D
	for _, spi := range sps {
		msp = &d.spans[spi]
		start, end = msp.start, msp.end
		if start < 0 {
			msp.start = 1
			msp.end = 1
			continue
		}
		sz = end - start
		if sz == 0 {
			continue
		}
		copy(D[top:], D[start:end])
		msp.start = top
		top += sz
		msp.end = top
	}
	d.free = append(d.free, d.rmq...)
	d.rmq = d.rmq[:0]
	d.D = D[:top]
	d.nFree = 0
}

func (d *T) dump() {
	fmt.Printf("to remove: %v\n", d.rmq)
	fmt.Printf("free: %v\n", d.free)
	for i := range d.spans {
		m := &d.spans[i]
		fmt.Printf("\t%03d %s: %v\n", i, m, d.D[m.start:m.end])
	}
}

func (d *T) mspan() (*mspan, Span) {
	N := len(d.free)
	if len(d.free) != 0 {
		spi := d.free[N-1]
		d.free = d.free[:N-1]
		sp := &d.spans[spi]
		return sp, spi
	}
	N = len(d.spans)
	if N == cap(d.spans) {
		M := N
		if M == 0 {
			M = 17
		}
		tmp := make([]mspan, M*5/3)
		copy(tmp, d.spans)
		d.spans = tmp
	}
	d.spans = d.spans[:N+1]
	res := &d.spans[N]
	return res, Span(N)
}
