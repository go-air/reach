package iic

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"sort"
	"time"

	"github.com/go-air/gini"
	"github.com/go-air/gini/logic"
	"github.com/go-air/gini/z"
	"github.com/go-air/reach/iic/internal/lits"
	"github.com/go-air/reach/iic/internal/queue"
)

// pre-processor for input aigs
type pp struct {
	lits    *lits.T
	sat     *gini.Gini
	trans   *logic.S
	bad     z.Lit
	orgLen  int
	clauses []clause
	freeIds []int
	occs    [][]int

	todo   *queue.T
	adders []z.Lit
	ssrAdd []z.Lit
	frozen []bool
	marks  []bool
	totry  []bool
	dcs    []int
	dms    []int

	verbose bool
}

func newPp(trans *logic.S, bad z.Lit) *pp {
	res := &pp{trans: trans, bad: bad, orgLen: trans.Len(), clauses: make([]clause, 1, 1024)}
	res.marks = make([]bool, trans.Len())
	res.totry = make([]bool, trans.Len())
	res.dcs = make([]int, trans.Len())
	res.dms = make([]int, trans.Len())
	for i := range res.totry {
		res.totry[i] = true
	}
	res.lits = lits.New()
	res.findFrozen()
	res.todo = queue.New(trans.Len())
	trans.ToCnf(res)
	return res
}

type clause struct {
	trans z.Lit
	ms    lits.Span
	sig   uint64
}

func (p *pp) processTo(sat *gini.Gini, deadLine *time.Time) {
	for {
		x, dc, dm := p.selectBestElim(deadLine)
		if dc > 1 || x == z.LitNull || (dc == 1 && dm > 10) {
			break
		}
		p.elim(x)
		p.ssr()
		if time.Until(*deadLine) <= 0 {
			break
		}
	}
	if p.verbose {
		fmt.Printf("done preprocessing.\n")
	}
	for i := range p.clauses {
		c := &p.clauses[i]
		if c.ms == 0 {
			continue
		}
		ms := p.lits.Get(c.ms)
		for _, m := range ms {
			sat.Add(m)
		}
		sat.Add(0)
	}
}

func (p *pp) ssr() bool {
	rez := &lits.Resolver{}
	var res = false
	var ms []z.Lit
	if p.verbose {
		fmt.Printf("pp ssr |q|=%d\n", p.todo.Len())
	}
	for p.todo.Len() > 0 {
		i := p.todo.Pop()
		c := &p.clauses[i]
		if c.ms == 0 {
			continue
		}
		if p.rmSubsumed(c, i) {
			res = true
		}
		ms = ms[:0]
		// NB addSsrs can change the memory backing p.lits
		ms = append(ms, p.lits.Get(c.ms)...)
		rez.Set(ms, 0)
		for _, m := range ms {
			p.addSsrs(c, i, rez, m)
		}
	}
	return res
}

func (p *pp) addSsrs(c *clause, id int, rez *lits.Resolver, m z.Lit) {
	occs := p.occs[m.Not()]
	if !rez.SetPivot(m.Var()) {
		log.Fatalf("rez %s set piv %s c%d %v\n", rez, m, id, p.lits.Get(c.ms))
	}
	var ms = p.lits.Get(c.ms)
	var N = len(ms)
	var oms []z.Lit
	var ok bool
	var rmq []int
	p.ssrAdd = p.ssrAdd[:0]
	for _, oid := range occs {
		oc := &p.clauses[oid]
		if oc.ms == 0 {
			continue
		}
		if oc.sig&c.sig != c.sig {
			continue
		}
		oms = p.lits.Get(oc.ms)
		if len(oms) < N {
			continue
		}
		if !lits.ContainedBySortedExcept(ms, oms, m) {
			continue
		}
		orgLen := len(p.ssrAdd)
		p.ssrAdd, ok = rez.Resolve(p.ssrAdd, oms)
		if !ok {
			panic("no resolve but contained by except")
		}
		if len(p.ssrAdd)-orgLen != len(oms)-1 {
			panic("wrong len resolve")
		}
		p.ssrAdd = append(p.ssrAdd, 0)
		rmq = append(rmq, oid)
	}
	for _, id := range rmq {
		p.remove(id)
	}
	n := 0
	for _, m := range p.ssrAdd {
		p.Add(m)
		if m == 0 {
			n++
		}
	}
	if n != 0 && debugPpSsr {
		fmt.Printf("add ssr %dc/%dm\n", n, len(p.ssrAdd)-n)
	}
	p.ssrAdd = p.ssrAdd[:0]
}

func (p *pp) rmSubsumed(c *clause, id int) bool {
	ms := p.lits.Get(c.ms)
	occs := p.occs
	minM := lits.MinLit(ms, func(m z.Lit) int {
		return len(occs[m])
	})
	rmq := make([]int, 0, 32)
	var oms []z.Lit
	for _, oid := range occs[minM] {
		oc := &p.clauses[oid]
		if oc.ms == 0 {
			continue
		}
		if oc.sig&c.sig != c.sig {
			continue
		}
		if oid == id {
			continue
		}
		oms = p.lits.Get(oc.ms)
		if len(oms) < len(ms) {
			continue
		}
		if !lits.ContainedBySorted(ms, oms) {
			continue
		}
		rmq = append(rmq, oid)
	}
	for _, id := range rmq {
		p.remove(id)
	}
	if p.verbose && len(rmq) != 0 {
		fmt.Printf("rmd %d clauses\n", len(rmq))
	}
	return len(rmq) > 0
}

func (p *pp) tryElim(m z.Lit) (dc, dm int) {
	pivot := m.Var()
	rez := &lits.Resolver{}
	var dst []z.Lit
	var ok bool
	//fmt.Printf("try elim %s:\n\toccs %v\n\t noccs %v\n", m, p.occs[m], p.occs[m.Not()])
	for _, id := range p.occs[m] {
		c := &p.clauses[id]
		if !rez.Set(p.lits.Get(c.ms), pivot) {
			panic("wilma!")
		}
		for _, oid := range p.occs[m.Not()] {
			oc := &p.clauses[oid]
			dst, ok = rez.Resolve(dst[:0], p.lits.Get(oc.ms))
			if !ok {
				continue
			}
			dc++
			dm += len(dst)
		}
	}
	for _, id := range p.occs[m] {
		c := &p.clauses[id]
		dc--
		dm -= p.lits.Len(c.ms)
	}
	for _, id := range p.occs[m.Not()] {
		c := &p.clauses[id]
		dc--
		dm -= p.lits.Len(c.ms)
	}
	return
}

func (p *pp) elim(m z.Lit) {
	if p.verbose {
		fmt.Printf("eiminating %s\n", m)
	}
	pivot := m.Var()
	rez := &lits.Resolver{}
	var dst []z.Lit
	var ok bool
	occs := p.occs[m]
	noccs := p.occs[m.Not()]
	for _, id := range occs {
		c := &p.clauses[id]
		if !rez.Set(p.lits.Get(c.ms), pivot) {
			panic("wilma!")
		}
		for _, oid := range noccs {
			oc := &p.clauses[oid]
			dst, ok = rez.Resolve(dst[:0], p.lits.Get(oc.ms))
			if !ok {
				continue
			}
			if debugPpElim {
				fmt.Printf("\tadd %v\n", dst)
			}
			for _, m := range dst {
				p.Add(m)
			}
			p.Add(0)
		}
	}
	moccs := p.occs[m]
	p.occs[m] = p.occs[m][:0]
	for _, id := range moccs {
		p.remove(id)
	}
	if debugPpElim {
		fmt.Printf("\trm %v\n", moccs)
	}

	moccs = p.occs[m.Not()]
	p.occs[m.Not()] = p.occs[m.Not()][:0]
	for _, id := range moccs {
		p.remove(id)
	}
	if debugPpSub {
		fmt.Printf("\trm %v\n", moccs)
	}
	p.frozen[pivot] = true
}

func (p *pp) findFrozen() {
	p.frozen = make([]bool, p.trans.Len())
	for _, m := range p.trans.Latches {
		p.frozen[m.Var()] = true
		n := p.trans.Next(m)
		p.frozen[n.Var()] = true
	}
	p.frozen[p.bad.Var()] = true
}

func (p *pp) selectRandElim() (z.Lit, int, int) {
	count := 0
	for count < 1024 {
		cand := z.Var(rand.Intn(p.trans.Len()-1)) + 1
		if p.frozen[cand] {
			count++
			continue
		}
		dc, dm := p.tryElim(cand.Pos())
		return cand.Pos(), dc, dm
	}
	return z.LitNull, 0, 0
}

func (p *pp) selectBestElim(deadLine *time.Time) (z.Lit, int, int) {
	dc := 1 << 30
	dm := 1 << 30
	var res z.Lit
	for i := range p.marks {
		p.marks[i] = false
	}
	for _, m := range p.trans.Latches {
		nxt := p.trans.Next(m)
		p.selectElimRec(p.marks, nxt, &dc, &dm, &res, deadLine)
	}
	p.selectElimRec(p.marks, p.bad, &dc, &dm, &res, deadLine)
	return res, dc, dm
}

func (p *pp) selectElimRec(marks []bool, m z.Lit, dcp, dmp *int, resp *z.Lit, deadLine *time.Time) {
	if marks[m.Var()] {
		return
	}
	marks[m.Var()] = true
	if p.trans.Type(m) == logic.SAnd {
		c0, c1 := p.trans.Ins(m)
		p.selectElimRec(marks, c0, dcp, dmp, resp, deadLine)
		p.selectElimRec(marks, c1, dcp, dmp, resp, deadLine)
	}
	if p.frozen[m.Var()] {
		return
	}
	if time.Until(*deadLine) <= 0 {
		return
	}
	var dc, dm int
	if !p.totry[m.Var()] {
		dc, dm = p.dcs[m.Var()], p.dms[m.Var()]
	} else {
		p.totry[m.Var()] = false
		dc, dm = p.tryElim(m)
		p.dcs[m.Var()], p.dms[m.Var()] = dc, dm
	}
	if dc < 0 {
		p.elim(m)
		return
	}
	if dc < *dcp || (dc == *dcp && dm < *dmp) {
		*dcp = dc
		*dmp = dm
		*resp = m
	} else if dc == *dcp && dm == *dmp {
		if rand.Intn(3) == 1 {
			*resp = m
		}
	}
}

func (p *pp) Add(m z.Lit) {
	if m == z.LitNull {
		// create clause
		cls, id := p.newClause()
		for _, m := range p.adders {
			p.ensureM(m)
			p.occs[m] = append(p.occs[m], id)
		}
		sort.Slice(p.adders, func(i, j int) bool {
			return p.adders[i] < p.adders[j]
		})
		cls.ms = p.lits.Put(p.adders)
		cls.sig = lits.CalcSig(p.adders)
		p.adders = p.adders[:0]
		p.todo.Push(id)
		return
	}
	p.adders = append(p.adders, m)
}

func (p *pp) remove(id int) {
	c := &p.clauses[id]
	ms := p.lits.Get(c.ms)
	for _, m := range ms {
		p.totry[m.Var()] = true
		occs := p.occs[m]
		j := 0
		for _, oid := range occs {
			if oid == id {
				continue
			}
			occs[j] = oid
			j++
		}
		p.occs[m] = occs[:j]
	}
	p.lits.Remove(c.ms)
	c.ms = 0
	p.freeIds = append(p.freeIds, id)
}

func (p *pp) newClause() (*clause, int) {
	if len(p.freeIds) != 0 {
		n := len(p.freeIds) - 1
		id := p.freeIds[n]
		p.freeIds = p.freeIds[:n]
		return &p.clauses[id], id
	}
	if len(p.clauses) == cap(p.clauses) {
		N := len(p.clauses)
		N += 7
		tmp := make([]clause, len(p.clauses), (N*5)/3)
		copy(tmp, p.clauses)
		p.clauses = tmp
	}
	N := len(p.clauses)
	p.clauses = p.clauses[:N+1]
	return &p.clauses[N], N
}

func (p *pp) ensureM(m z.Lit) {
	im := int(m) + 1
	if im < len(p.occs) {
		return
	}
	tmp := make([][]int, im*2)
	copy(tmp, p.occs)
	p.occs = tmp
}

func (p *pp) dump(dst io.Writer) {
	for i := range p.clauses[1:] {
		c := &p.clauses[i+1]
		if c.ms == 0 {
			continue
		}
		for _, m := range p.lits.Get(c.ms) {
			fmt.Fprintf(dst, "%s ", m)
		}
		fmt.Fprintf(dst, "0\n")
	}
}

func (p *pp) numClauses() int {
	return len(p.clauses) - len(p.freeIds) - 1
}
