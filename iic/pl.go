package iic

import (
	"fmt"
	"os"

	"github.com/go-air/gini/z"
	"github.com/go-air/reach"
	"github.com/go-air/reach/iic/internal/cnf"
	"github.com/go-air/reach/iic/internal/obs"
)

type pl struct {
	sifter *sifter

	k       int
	nBlock  int
	nExtend int

	lastSiftLen  int
	lastnBlock   int64
	sifts        int64
	siftAttempts int64
	siftReduced  int64
}

func (l *pl) crmHook(c cnf.Id) {
}

func (l *pl) prop(f *cnf.T, sat *satmon, bad z.Lit, pri *reach.Primer, obs *obs.Set) (timeOk bool) {
	timeOk = true
	if debugPushes {
		fmt.Printf("prop level %d len %d\n", l.k, f.LenK(l.k))
	}
	f.RemoveDups(l.k)
	if debugPushes {
		fmt.Printf("\t remove dups len %d\n", f.LenK(l.k))
	}

	K := f.K()
	sat.Assume(bad.Not())
	for k := l.k; k <= K; k++ {
		f.Forall(k, func(f *cnf.T, c cnf.Id) {
			sat.Assume(f.ActLit(c))
		})
	}
	defer func() {
		sat.Untest()
		f.Simplify(l.k)
	}()
	if st, _ := sat.Test(nil); st != 0 {
		//panic("prop test solved")
	}
	f.Forall(l.k, func(f *cnf.T, c cnf.Id) {
		if !timeOk {
			return
		}
		switch l.tryPush(f, c, sat, pri) {
		case 0:
			timeOk = false
		case 1:
			if debugPushLevel {
				fmt.Printf("\tnot pushing %s\n", f.String(c))
			}
		case -1:
			if debugPushLevel {
				fmt.Printf("\tpushing %s\n", f.String(c))
			}
			f.Push(c)
			obs.BlockAt(l.k+1, f.Lits(c))
		}
	})
	if debugPushes {
		fmt.Printf("done prop level %d len %d\n", l.k, f.LenK(l.k))
	}
	return
}

func (l *pl) conSift(f *cnf.T, sat *satmon, obs *obs.Set, init, bad z.Lit) (timeOk bool) {
	timeOk = true
	l.sifts++
	start := l.k - 1
	if start == 0 {
		sat.Assume(init)
		start++
	}
	f.RemoveDups(l.k)
	if debugConSift {
		fmt.Printf("begin sift level %d\n", l.k)
		f.Stats(os.Stdout)
	}
	K := f.K()
	sat.Assume(bad.Not())
	for k := start; k <= K; k++ {
		f.Forall(k, func(f *cnf.T, c cnf.Id) {
			sat.Assume(f.ActLit(c))
		})
	}
	sat.Assume(bad.Not())

	if st, _ := sat.sat.Test(nil); st != 0 {
		panic("sift test solved")
	}
	toAdd := make([]z.Lit, 0, 32)
	defer func() {
		sat.sat.Untest()
		if len(toAdd) == 0 {
			return
		}
		l.addToAdd(f, obs, toAdd)
		f.Simplify(l.k)
		if debugConSift {
			fmt.Printf("end sift level %d\n", l.k)
			f.Stats(os.Stdout)
		}
	}()
	f.Forall(l.k, func(f *cnf.T, c cnf.Id) {
		if !timeOk {
			return
		}
		orgLen := len(toAdd)
		toAdd, timeOk = l.sifter.sift(toAdd, c)
		l.siftAttempts++
		if orgLen != len(toAdd) {
			l.siftReduced++
		}
	})
	return timeOk
}

func (l *pl) addToAdd(f *cnf.T, obs *obs.Set, toAdd []z.Lit) {
	last := 0
	N := len(toAdd)
	var ms []z.Lit
	for i := 1; i < N; i++ {
		if toAdd[i] != 0 {
			continue
		}
		ms = toAdd[last:i]
		f.Add(ms, l.k)
		obs.BlockAt(l.k, ms)
		i++
		last = i
	}
}

func (l *pl) tryPush(cnf *cnf.T, c cnf.Id, sat *satmon, pri *reach.Primer) int {
	ms := cnf.Lits(c)
	for _, m := range ms {
		mp := pri.Prime(m)
		sat.Assume(mp.Not())
	}
	return sat.Try()
}
