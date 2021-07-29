// Copyright 2018 Scott Cotton. All rights reserved.  Use of this source
// code is governed by a license that can be found in the License file.

package iic

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/go-air/reach/iic/internal/lits"
	"github.com/go-air/reach/iic/internal/obs"

	"github.com/go-air/gini"
	"github.com/go-air/gini/logic"
	"github.com/go-air/gini/z"
	"github.com/go-air/reach"
	"github.com/go-air/reach/iic/internal/cnf"
)

// T contains state for an ic3/pdr model checker.
type T struct {
	trans       *logic.S
	orgTransLen int
	init        z.Lit
	bad         z.Lit // bad state over latches
	badPrime    z.Lit // bad state over next states of latches
	rResult     *reach.Result
	sat         *gini.Gini
	lits        *lits.T
	cnf         *cnf.T
	obs         *obs.Set
	gnrl        *gnrl
	pushes      *np
	preproc     *pp

	blkSat    *satmon
	propSat   *satmon
	gnrlSat   *satmon
	primer    *reach.Primer
	justifier *justifier

	opts *Options

	maxDepth  int
	deadLine  time.Time
	startTime time.Time
	blockTime time.Duration
	traceHd   obs.Id
	mps       []z.Lit // scratch primes to justify
	initVals  []int8
	learnts   int64
}

// New creates a new incremental inductive model checker from a transition
// system and bad state literal.
func New(trans *logic.S, bad z.Lit) *T {
	init := trans.T
	initVals := make([]int8, trans.Len())
	if debugState {
		fmt.Printf("init values:\n")
	}
	for _, m := range trans.Latches {
		switch trans.Init(m) {
		case trans.T:
			if debugState {
				fmt.Printf("%s: true\n", m)
			}
			init = trans.And(init, m)
			initVals[m.Var()] = 1
		case trans.F:
			if debugState {
				fmt.Printf("%s: false\n", m)
			}
			init = trans.And(init, m.Not())
			initVals[m.Var()] = -1
		default:
			if debugState {
				fmt.Printf("%s: x\n", m)
			}
		}
		trans.SetInit(m, z.LitNull)
	}
	res := &T{trans: trans, init: init, bad: bad, orgTransLen: trans.Len(),
		sat: gini.NewVc(trans.Len()+16384, trans.Len()+16384)}
	res.lits = lits.New()
	res.initVals = initVals
	res.cnf = cnf.New(res.sat, res.lits)
	// does CNF as well
	res.primer = reach.NewPrimer(trans, init, bad)
	trans.ToCnf(res.sat)
	res.badPrime = res.prime(bad)
	res.obs = obs.NewSet(res.lits)
	res.opts = NewOptions()
	res.blkSat = newSatMon("block", res.sat, &res.deadLine)
	res.propSat = newSatMon("prop", res.sat, &res.deadLine)
	res.gnrlSat = newSatMon("gnrl", res.sat, &res.deadLine)
	res.gnrl = newGnrl(res.gnrlSat, trans, res.lits, res.obs, res.initVals)
	res.justifier = newJustifier(trans)
	res.pushes = newNp(res.cnf, res.propSat, res.primer, res.obs, res.initVals, res.init, res.bad)
	res.preproc = newPp(res.trans, res.bad)
	res.rResult = &reach.Result{M: res.bad}
	res.maxDepth = 1 << 30
	res.cnf.SetRemoveHook(func(f *cnf.T, c, by cnf.Id, k int) {
		res.pushes.crmHook(f, c, by, k)
	})
	res.maxDepth = 1 << 30
	return res
}

func (t *T) Options() *Options {
	return t.opts
}

func (t *T) installOpts() {
	t.preproc.verbose = t.opts.Verbose
	if t.opts.DeepObs {
		obs.Requeue = obs.RequeueLong
	} else {
		obs.Requeue = obs.RequeueShort
	}
	t.obs.FilterBlocked = t.opts.FilterObs
	t.startTime = time.Now()
	t.deadLine = t.startTime.Add(t.opts.Duration)
	t.gnrl.doRemoveLits = t.opts.GnrlRemoveLits
	t.pushes.conSift = t.opts.ConsecuSift
	t.pushes.conSiftPull = t.opts.ConsecuSiftPull
}

// Try tries to solve the reachability problem specified in New.
//
// Try returns
//
//  1 if there is a trace to the bad state
//  0 if timed out
//  -1 if there cannot be a trace to the bad state
func (t *T) Try() int {
	t.installOpts()
	if t.opts.Preprocess {
		t.preproc.processTo(t.sat, &t.deadLine)
	} else {
		t.trans.ToCnf(t.sat)
	}
	if res := t.ckInit(); res != -1 {
		return res
	}
	t.rResult.Depth = 1
	t.cnf.PushK()
	t.cnf.PushK()
	t.pushes.push()
	t.pushes.push()
	if t.opts.Verbose {
		defer t.stats()
	}
	ticker := time.NewTicker(time.Second / 2)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if t.opts.Verbose {
				t.stats()
			}
		default:
		}
		ob := t.obs.Choose()
		if ob == 0 {
			K := t.obs.MaxK() + 1
			if K > t.maxDepth {
				return 0
			}
			t.cnf.PushK()
			fixedPoint, timeOk := t.pushes.prop(K)
			if !timeOk {
				return 0
			}
			if fixedPoint {
				t.rResult.SetUnreachable()
				if err := t.verifyInd(); err != nil {
					log.Fatalf("error verifying invariable: %s\n", err)
				}
				t.pushes.push()
				return -1
			}
			t.rResult.Depth = K
			if t.opts.Verbose {
				fmt.Printf("increase K to %d\n", K)
				t.stats()
			}

			t.obs.Grow()
			t.pushes.push()
			continue
		}
		if debugObq {
			fmt.Printf("[obq]: chose %s\n", t.obs.String(ob))
		}
		// t.pushes.preHandleOb(t.obs.K(ob))
		res, nob := t.handleOb(ob)
		switch res {
		case obBlocked:
			if debugObq {
				fmt.Printf("[obq]: blocked %s\n", t.obs.String(ob))
			}
		case obExtended:
			if debugObq {
				fmt.Printf("[obq]: extend.\n")
			}
			if t.obs.InitWit(nob) == z.LitNull || t.obs.K(nob) == 0 {
				t.traceHd = nob
				t.rResult.Depth = t.obs.DistToBad(nob)
				t.rResult.SetReachable(nil)
				return 1
			}
		case obTimeout:
			if debugObq {
				fmt.Printf("[obq]: timeout.\n")
			}
			return 0
		default:
			panic(fmt.Sprintf("unknown obres: %s", res))
		}
	}
}

func (t *T) handleOb(o obs.Id) (res obRes, nob obs.Id) {
	if debugState || debugHandle {
		fmt.Printf("begin handleOb %s\n", t.obs.String(o))
	}
	if t.obs.IsRoot(o) {
		return t.handleRootBlockOrExtend(o)
	}
	return t.handleObBlockOrExtend(o)
}

func (t *T) handleObBlockOrExtend(o obs.Id) (obRes, obs.Id) {
	if debugHandle {
		fmt.Printf("handleObBlockOrExtend: %s\n", t.obs.String(o))
	}
	for _, m := range t.obs.Ms(o) {
		t.sat.Add(m.Not())
	}
	act := t.sat.Activate()
	defer t.sat.Deactivate(act)
	t.assumeLevel(t.obs.K(o) - 1)
	t.sat.Assume(act)
	t.sat.Assume(t.bad.Not())
	if res, _ := t.sat.Test(nil); res == -1 {
		// NB untest in handleObIndGnr
		t.handleObIndGnr(o)
		return obBlocked, 0
	}
	t.assumePrimes(t.obs.Ms(o))
	st := t.blockCallSat()
	if st == 0 {
		if st := t.sat.Untest(); st == -1 {
			panic("untest handle ob sat")
		}
		return obTimeout, 0
	}
	if st == -1 {
		// NB untest in handleObIndGnr
		t.handleObIndGnr(o)
		return obBlocked, 0
	}
	if debugState {
		fmt.Printf("state @ %d:\n", t.obs.K(o)-1)
		for _, m := range t.trans.Latches {
			nxt := t.trans.Next(m)
			fmt.Printf("%s %t (nxt %s %s %t)\n", m, t.sat.Value(m),
				nxt, t.trans.Type(nxt), t.sat.Value(nxt))
		}
	}
	res, nob := t.extend(o, t.obs.Ms(o))
	if res := t.sat.Untest(); res == -1 {
		panic("untest handle ob unsat")
	}
	return res, nob
}

// after sat.
func (t *T) extend(o obs.Id, mps []z.Lit) (obRes, obs.Id) {
	var ms []z.Lit
	if t.opts.Justify {
		// nb justify uses sat as model, can't untest before
		// this is done
		ms = t.justifyPrimes(nil, mps)
	} else {
		ms = make([]z.Lit, len(t.trans.Latches))
		for i, m := range t.trans.Latches {
			if t.sat.Value(m) {
				ms[i] = m
			} else {
				ms[i] = m.Not()
			}
		}
	}
	ini := t.findInit()
	if len(ms) == 0 {
		// yes this can happen
		//log.Fatalf("zero j of %v\n", t.obs.ms(o))
	}
	nob := t.obs.Extend(o, ms, ini)
	if debugState {
		fmt.Printf("\tnob %s <- %s\n", t.obs.String(nob), t.obs.String(o))
	}
	return obExtended, nob
}

func (t *T) handleRootBlockOrExtend(o obs.Id) (obRes, obs.Id) {
	if debugHandle {
		fmt.Printf("handleRootBlockOrExtend %s\n", t.obs.String(o))
	}
	t.sat.Assume(t.badPrime)
	t.sat.Assume(t.bad.Not())
	t.assumeLevel(t.obs.K(o) - 1)
	res := t.blockCallSat()
	switch res {
	case -1:
		// main loop handles root specially here,
		//
		// since it has uniqu max depth > cnf.K()
		// this willl trigger propTo in t.Try()
		// which in turn identifies when an inductive invariant
		// is found.
		return obBlocked, 0
	case 1:
		return t.extend(o, []z.Lit{t.bad})
	case 0:
		return obTimeout, 0
	}
	panic("unreachable")
}

func (t *T) handleObIndGnr(o obs.Id) {
	if debugHandle {
		fmt.Printf("handle ind->gnrl\n")
	}
	if !t.gnrl.gnrlize(o, t.primer) {
		return // time is up
	}
	t.learnts++
	ms, _ := t.gnrl.cnfMs()
	k := t.obs.K(o)
	K := t.cnf.K()
	c := t.cnf.Add(ms, k)
	act := t.cnf.ActLit(c)
	t.obs.Block(o, ms)
	for configPushOnBlock && k < K {
		k++
		t.sat.Assume(act)
		t.cnf.AssumeLevel(k)
		for _, m := range ms {
			t.sat.Assume(t.primer.Prime(m).Not())
		}
		res := t.sat.Try(time.Until(t.deadLine))
		switch res {
		case 0:
			k = K
		case 1:
			k = K
		case -1:
			t.obs.BlockAt(k, ms)
			t.cnf.Push(c)
		}

	}
	t.pushes.onBlock(c)
	if debugLearn {
		fmt.Printf("learn block %s with %s\n", t.obs.String(o), t.cnf.String(c))
	}
}

// find a latch whose current sat valuation violates init condition.
// when this is true, and we block the state, then the resulting blocking
// clause will be satisfiable by all initial states.  When this is false,
// we have found a trace.
func (t *T) findInit() z.Lit {
	var res z.Lit
	for _, m := range t.trans.Latches {
		if !t.sat.Value(m) {
			m = m.Not()
		}
		if t.initVals[m.Var()]+m.Sign() == 0 {
			res = m
			break
		}
	}
	if debugInits {
		fmt.Printf("[inits] witness is %s\n", res)
	}
	return res
}

// check if (1) I -> P, if not trace check if (2) I & T -> P', if not trace
//
// if trace, record info necessary to generate traceHd member.
// return 1 if either (1) or (2) doesn't hold
// 0 if timeout.
// -1 if (1) and (2) hold
func (t *T) ckInit() int {
	// TBD: make a trace ok to generate from sat.
	t.sat.Assume(t.init)
	t.sat.Assume(t.bad)
	res := t.callSat()
	switch res {
	case 0:
		return 0
	case 1:
		t.rResult.SetReachable(nil)
		ms := make([]z.Lit, 0, len(t.trans.Latches))
		for _, m := range t.trans.Latches {
			if !t.sat.Value(m) {
				m = m.Not()
			}
			ms = append(ms, m)
		}
		t.traceHd = t.obs.Extend(t.obs.Root(), ms, 0)
		return 1
	}
	t.sat.Assume(t.init)
	t.sat.Assume(t.bad.Not())
	t.sat.Assume(t.badPrime)
	res = t.callSat()
	if res != 1 {
		return res
	}
	t.rResult.SetReachable(nil)
	ms := make([]z.Lit, 0, len(t.trans.Latches))
	ns := make([]z.Lit, 0, len(t.trans.Latches))
	for _, m := range t.trans.Latches {
		mp := t.prime(m)
		if !t.sat.Value(mp) {
			mp = mp.Not()
		}
		ns = append(ns, mp)
		if !t.sat.Value(m) {
			m = m.Not()
		}
		ms = append(ms, m)
	}
	p := t.obs.Extend(t.obs.Root(), ns, z.LitNull)
	t.traceHd = t.obs.Extend(p, ms, z.LitNull)
	return 1
}

func (t *T) verifyInd() error {
	K := t.cnf.K()
	if debugVerifyInd {
		fmt.Printf("verifying inductive result at depth %d.\n", K)
	}
	t.sat.Assume(t.bad.Not())
	t.cnf.AssumeLevel(K)
	defer t.sat.Untest()

	if res, _ := t.sat.Test(nil); res == -1 {
		return fmt.Errorf("ErrIndCurLevelUnsat")
	}
	t.sat.Assume(t.badPrime)
	switch st := t.sat.Solve(); st {
	case 1:
		return fmt.Errorf("not(bad) is not inductive")
	case -1:
		if debugVerifyInd {
			fmt.Printf("verified that not(bad) is inductive.\n")
		}
	}

	var err error
	t.cnf.Forall(K, func(f *cnf.T, c cnf.Id) {
		if err != nil {
			return
		}
		if debugVerifyInd {
			fmt.Printf("checking consecution for %s\n", c)
		}
		ms := f.Lits(c)
		for _, m := range ms {
			mp := t.prime(m)
			t.sat.Assume(mp.Not())
		}
		switch st := t.sat.Solve(); st {
		case 1:
			err = fmt.Errorf("clause %v not consecutive", ms)
		case 0:
			err = fmt.Errorf("timeout during sanity check verifying invariant")
		}
	})
	// TBD: initiation.
	return nil
}

// FillOutput fills `o` with information about the last
// call to Try.
func (t *T) FillOutput(o *reach.Output) {
	t.rResult.Dur = time.Since(t.startTime)
	if t.rResult.IsUnreachable() {
		t.cnf.Simplify(t.cnf.K())
		t.cnf.Forall(t.cnf.K(), func(f *cnf.T, c cnf.Id) {
			for _, m := range t.cnf.Lits(c) {
				t.rResult.Add(m)
			}
			t.rResult.Add(0)
		})
		t.rResult.Add(t.bad.Not())
		t.rResult.Add(0)
	} else if t.rResult.IsReachable() {
		tr, terr := t.buildTrace()
		if terr != nil {
			log.Printf("error generating trace: %s", terr)
		}
		t.rResult.Trace = tr
	}
	o.AppendResult(t.rResult)
}

func (t *T) buildTrace() (*reach.Trace, error) {
	tg := &traceGen{
		trans:    t.trans,
		orgLen:   t.orgTransLen,
		primer:   t.primer,
		init:     t.init,
		bad:      t.bad,
		badPrime: t.badPrime,
		hd:       t.traceHd,
		obs:      t.obs,
		sat:      newSatMon("tracegen", t.sat, &t.deadLine)}
	return tg.build()
}

func (t *T) stats() {
	t.blkSat.Stats(os.Stdout)
	t.propSat.Stats(os.Stdout)
	t.gnrl.Stats(os.Stdout)
	t.cnf.Stats(os.Stdout)
	t.pushes.Stats(os.Stdout)
}

func (t *T) assumePrimes(ms []z.Lit) {
	for _, m := range ms {
		mp := t.prime(m)
		t.blkSat.Assume(mp)
	}
}

func (t *T) justifyPrimes(dst, ms []z.Lit) []z.Lit {
	sat := t.sat
	justifier := t.justifier
	justifier.JustifyInit()
	for _, m := range ms {
		mp := t.prime(m)
		dst = justifier.JustifyOne(dst, sat, mp)
	}
	return dst
}

func (t *T) assumeLevel(k int) {
	if k == 0 {
		t.blkSat.Assume(t.init)
		return
	}
	t.cnf.AssumeLevel(k)
}

func (t *T) prime(m z.Lit) z.Lit {
	return t.primer.Prime(m)
}

func (t *T) curLevel() int {
	return t.cnf.K() - 1
}

func (t *T) blockCallSat() int {
	return t.blkSat.Try()
}

func (t *T) callSat() int {
	dur := time.Until(t.deadLine)
	if dur < 0 {
		return 0
	}
	return t.sat.Try(dur)
}
