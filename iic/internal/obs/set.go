package obs

import (
	"container/heap"
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/go-air/gini/z"
	"github.com/go-air/reach/iic/internal/lits"
)

const (
	debug = false
)

// Set represents a set of proof obligations.
type Set struct {
	FilterBlocked bool

	lits *lits.T
	d    []ob
	ks   [][]Id
	next []Id
	free []Id

	kOccs int
	occs  [][]Id // only for k=kOccs

	// crazy state
	k     int // current level
	kStar int // max level of non-root
}

// NewSet creates a new set of proof obligations
// backed by d.
func NewSet(d *lits.T) *Set {
	res := &Set{
		lits: d,
		d:    make([]ob, 1, 128)}
	root, rid := res.newOb()
	res.ensureLevel(2)
	res.k = 1
	res.kOccs = 0
	res.kStar = 2
	res.FilterBlocked = true
	root.k = 2
	root.chosen = true
	res.ks[2] = []Id{rid}
	return res
}

// Root returns the root Obligation
func (s *Set) Root() Id {
	return Id(1)
}

// IsRoot returns whether o is the root proof obligation.
func (s *Set) IsRoot(o Id) bool {
	return o == Id(1)
}

// Extends extends parent with a new proof obligation that
// can reach parent in 1 step.
//
// k - the level of the resulting ob
// d - the distance to the bad state
// parent - the next step to the bad state
// ms - a justification of the step over state variables
// ini - a witness to not being in the initial state.
//
// return the id of a new proof obligation with the properties above.
func (s *Set) Extend(parent Id, ms []z.Lit, ini z.Lit) Id {
	sort.Slice(ms, func(i, j int) bool {
		return ms[i] < ms[j]
	})
	for _, m := range ms {
		s.ensureM(m)
	}
	pob := &s.d[parent]
	k, d := pob.k-1, pob.distToBad+1
	pob.nKids++
	heap.Push(s.asHeap(), parent)
	ob, id := s.newOb()
	ob.k = k
	ob.distToBad = d
	ob.parent = parent
	ob.ms = s.lits.Put(ms)
	ob.initWitness = ini
	ob.sig = lits.CalcSig(ms)
	ob.chosen = true
	ob.nKids = 0
	s.k = k
	heap.Push(s.asHeap(), id)
	return id
}

// returns a proof obligation to try to extend or block.
//
// If none available returns 0.  Then the user may call s.Grow()
//
func (s *Set) Choose() Id {
	if debug {
		s.Dump(os.Stdout)
	}
	for s.k < s.kStar {
		obq := s.asHeap()
		for obq.Len() > 0 {
			id := heap.Pop(obq).(Id)
			if s.K(id) != s.k {
				continue
			}
			s.d[id].chosen = true
			return id
		}
		s.k++
	}
	obq := s.asHeap()
	if obq.Len() > 0 {
		id := heap.Pop(obq).(Id)
		if s.K(id) != s.k {
			dPrintf("zero cause root pushed: %d %d\n", s.k, s.K(id))
			return Id(0)
		}
		if id != s.Root() {
			panic("wilma!")
		}
		s.d[id].chosen = true
		return id
	}
	dPrintf("zero cause nothing there.\n")
	return Id(0)
}

func (s *Set) Grow() {
	if debug {
		fmt.Printf("[obq]: grow |next| is %d\n", len(s.next))
	}
	s.kStar++
	s.ensureLevel(s.kStar)
	s.push(s.Root())

	n := 0
	for _, nxt := range s.next {
		if s.K(nxt) != s.kStar-1 {
			continue
		}
		n++
		s.ks[s.MaxK()] = append(s.ks[s.MaxK()], nxt)
	}
	if debug {
		fmt.Printf("[obq]: grow |ok(next)| is %d\n", n)
	}
	//s.ks[s.MaxK()] = append(s.ks[s.MaxK()], s.next...)
	s.next = s.next[:0]
	orgK := s.k
	s.k = s.MaxK()
	heap.Init(s.asHeap())
	s.k = orgK
}

func (s *Set) MaxK() int {
	return s.kStar - 1
}

func (s *Set) Parent(o Id) Id {
	return s.d[o].parent
}

func (s *Set) Block(o Id, ms []z.Lit) {
	if o == s.Root() {
		panic("can't block root.")
	}
	ob := &s.d[o]
	at := ob.k
	if Requeue(ob.k, ob.distToBad, s.MaxK()) {
		s.push(o)
	} else {
		ob.k++
		s.next = append(s.next, o)
	}

	s.BlockAt(at, lits.Flip(ms))
	lits.Flip(ms)
}

func (s *Set) BlockAt(k int, ms []z.Lit) {
	if !s.FilterBlocked {
		return
	}
	if s.kOccs != s.k {
		s.occs = s.buildOccs()
		s.kOccs = s.k
	}

	s.filt(ms, k)
	s.kOccs = -1
}

func (s *Set) buildOccs() [][]Id {
	occs := s.occs
	for i := range occs {
		occs[i] = occs[i][:0]
	}
	for _, id := range s.ks[s.k] {
		ob := &s.d[id]
		if ob.k != s.k {
			continue
		}
		for _, m := range s.lits.Get(ob.ms) {
			occs[m] = append(occs[m], id)
		}
	}
	return occs
}

func dPrintf(fmtString string, args ...interface{}) {
	if debug {
		fmt.Printf(fmtString, args...)
	}
}

func (s *Set) filt(ms []z.Lit, k int) {
	for _, m := range ms {
		s.ensureM(m)
	}
	m := lits.MinLit(ms, func(m z.Lit) int {
		return len(s.occs[m])
	})
	sig := lits.CalcSig(ms)
	j := 0
	sl := s.occs[m]
	D := s.lits
	maxK := s.MaxK()

	dPrintf("minLit: %s\n", m)
	for _, id := range s.occs[m] {
		ob := &s.d[id]
		dPrintf("visit %s... ", id)
		// filter these
		if ob.k != s.k {
			dPrintf("k stale, skipping\n")
			continue
		}
		// keep these at this level
		if ob.sig&sig != sig {
			dPrintf("sig mismatch, keeping\n")
			sl[j] = id
			j++
			continue
		}
		oms := D.Get(ob.ms)
		if len(oms) < len(ms) {
			sl[j] = id
			j++
			continue
		}
		if !lits.ContainedBySorted(ms, oms) {
			dPrintf("not contained by, keeping\n")
			sl[j] = id
			j++
			continue
		}
		if !Requeue(ob.k, ob.distToBad, maxK) {
			dPrintf("don't requeue. skipping\n")
			if ob.nKids == 0 && !ob.chosen {
				s.freeOb(id, ob)
			} else {
				ob.chosen = false
				s.next = append(s.next, id)
				ob.k++
			}
			continue
		}
		dPrintf("pushing.\n")
		// push these
		s.push(id)
	}
	dPrintf("done block filt\n")
	s.occs[m] = sl[:j]
}

func (s *Set) freeOb(id Id, o *ob) {
	dPrintf("freeing %s\n", s.String(id))
	s.lits.Remove(o.ms)
	o.ms = 0
	o.k = -1
	op := &s.d[o.parent]
	op.nKids--
	o.parent = 0
	s.free = append(s.free, id)
}

func (s *Set) String(o Id) string {
	return fmt.Sprintf("%s: %s %v", o, &s.d[o], s.lits.Get(s.d[o].ms))
}

func (s *Set) push(o Id) {
	ob := &s.d[o]
	if !s.IsRoot(o) && ob.k != s.k {
		panic(fmt.Sprintf("wilma!, ob.k is %d while s.k is %d", ob.k, s.k))
	}
	ob.k++
	orgK := s.k
	s.k = ob.k
	heap.Push(s.asHeap(), o)
	s.k = orgK
}

func (s *Set) K(o Id) int {
	return s.d[o].k
}

func (s *Set) DistToBad(o Id) int {
	return s.d[o].distToBad
}

func (s *Set) InitWit(o Id) z.Lit {
	return s.d[o].initWitness
}

func (s *Set) Ms(o Id) []z.Lit {
	return s.lits.Get(s.d[o].ms)
}

func (s *Set) Dump(dst io.Writer) {
	for k := range s.ks {
		fmt.Fprintf(dst, "obs level %d %d obs:\n", k, len(s.ks[k]))
		for _, id := range s.ks[k] {
			fmt.Fprintf(dst, "\t%s\n", s.String(id))
		}
	}
}

func (s *Set) ensureLevel(k int) {
	for k >= len(s.ks) {
		s.ks = append(s.ks, []Id{})
	}
}

func (s *Set) ensureM(m z.Lit) {
	im := int(m)
	for im >= len(s.occs) {
		s.occs = append(s.occs, []Id{})
	}
}

func (s *Set) asHeap() heap.Interface {
	return (*obq)(s)
}

func (s *Set) newOb() (*ob, Id) {
	N := len(s.free)
	if N > 0 {
		N--
		id := s.free[N]
		s.free = s.free[:N]
		return &s.d[id], id
	}
	N = len(s.d)
	if N == cap(s.d) {
		tmp := make([]ob, N*5/3)
		copy(tmp, s.d)
		s.d = tmp
	}
	s.d = s.d[:N+1]
	ob := &s.d[N]
	return ob, Id(N)
}
