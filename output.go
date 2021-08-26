// Copyright 2018 The Reach Authors. All rights reserved.  Use of this source
// code is governed by a license that can be found in the License file.

package reach

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-air/gini"
	"github.com/go-air/gini/dimacs"
	"github.com/go-air/gini/inter"
	"github.com/go-air/gini/logic"
	"github.com/go-air/gini/logic/aiger"
	"github.com/go-air/gini/z"
)

const (
	aigName  = "aig"
	traceExt = ".trace"
	invExt   = "-inv.cnf"
	badExt   = "-bad.json"
)

// Output encapsulates the output of the reach command
// checking subcommands.
type Output struct {
	root     string
	bads     []*Result
	deadline time.Time // for time limiting verification of results.
}

// MakeOutput creates an output object backed by directory root
// dir and binary aiger file with path g.
//
// MakeOutput tries to create dir/basename(g) as a directory
// where basename(g) is the filename of g (not the full path)
// with the extension removed.
//
// MakeOutput then tries to symlink g into the dir above.
// If all that goes well, MakeOutput returns an output object
// and a nil error.  Otherwise, MakeOutput returns a nil Output
// and a non-nil error.
func MakeOutput(g, dir string) (*Output, error) {
	var err error
	g, err = filepath.Abs(g)
	if err != nil {
		return nil, err
	}
	base := filepath.Base(g)
	base = strings.TrimSuffix(base, filepath.Ext(base))
	root := filepath.Join(dir, base)
	_, e := os.Stat(root)
	if e == nil {
		return nil, os.ErrExist
	}
	if e := os.MkdirAll(root, 0755); e != nil {
		return nil, e
	}
	res := &Output{root: root}
	if err := os.Symlink(g, res.AigerPath()); err != nil {
		return nil, err
	}
	return res, nil
}

// OpenOutput tries to open an output as created by MakeOutput.
func OpenOutput(d string) (*Output, error) {
	out := &Output{root: d}
	if err := out.readResults(); err != nil {
		return nil, err
	}
	return out, nil
}

// Results returns the bad states for which `o` contains
// result information.
func (o *Output) Results() []*Result {
	return o.bads
}

func (o *Output) readResults() error {
	fis, err := ioutil.ReadDir(o.root)
	if err != nil {
		return err
	}
	for _, fi := range fis {
		if fi.IsDir() {
			continue
		}
		if !strings.HasSuffix(fi.Name(), badExt) {
			continue
		}
		if err := o.readResult(fi.Name()); err != nil {
			return err
		}
	}
	return nil
}

func (o *Output) readResult(fn string) error {
	p := filepath.Join(o.root, fn)
	f, e := os.Open(p)
	if e != nil {
		return e
	}
	defer f.Close()
	bs, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}
	bad := &Result{}
	if err := json.Unmarshal(bs, bad); err != nil {
		return err
	}
	o.bads = append(o.bads, bad)
	return nil
}

// IsVerifiable returns whether or not the `i`th bad
// states formula has either
//   1. a trace and is reachable; or
//   2. an invariant is unreachable
//
// IsVerifiable checks the existence of files by
// os.Stat to accomplish this.
func (o *Output) IsVerifiable(i int) bool {
	b := o.bads[i]
	if !b.IsSolved() {
		return false
	}
	if b.IsReachable() && b.Trace != nil {
		return true
	}
	if !b.IsReachable() && len(b.Invariant) > 0 {
		return true
	}
	if b.IsReachable() {
		_, err := os.Stat(o.TracePath(i))
		if err != nil {
			return false
		}
		return true
	}
	_, err := os.Stat(o.InvariantPath(i))
	return err == nil
}

// AppendResult lets a checker append bad state information
// to the output.
func (o *Output) AppendResult(bads ...*Result) {
	o.bads = append(o.bads, bads...)
}

// Store attempts to store `o`, including any traces or
// invariants found in it's bad states.  Store returns
// a non-nil error if there is a problem doing this.
func (o *Output) Store() error {
	for i := range o.bads {
		if err := o.storeResult(i); err != nil {
			return err
		}
	}
	return nil
}

func (o *Output) storeResult(i int) error {
	bad := o.bads[i]
	badPath := o.ResultPath(i)
	f, e := os.Create(badPath)
	if e != nil {
		return e
	}
	defer f.Close()
	d, e := json.MarshalIndent(bad, "", "\t")
	if e != nil {
		return e
	}
	if _, e := f.Write(d); e != nil {
		return e
	}
	if bad.Trace != nil {
		if !bad.IsSolved() || !bad.IsReachable() {
			panic(fmt.Sprintf("bad bad: %s", bad))
		}
		p := o.TracePath(i)
		f, e := os.Create(p)
		if e != nil {
			return e
		}
		defer f.Close()
		if err := bad.Trace.Encode(f); err != nil {
			return err
		}
	}
	if len(bad.Invariant) != 0 {
		if !bad.IsSolved() || bad.IsReachable() {
			panic(fmt.Sprintf("bad bad: %s", bad))
		}
		return o.writeInv(i)
	}

	return nil
}

func (o *Output) writeInv(i int) error {
	bad := o.bads[i]
	p := o.InvariantPath(i)
	f, e := os.Create(p)
	if e != nil {
		return e
	}
	defer f.Close()
	nv, nc := dimacsHeader(bad)
	if _, err := fmt.Fprintf(f, "p cnf %d %d\n", nv, nc); err != nil {
		return err
	}
	var err error
	for _, m := range bad.Invariant {
		if m == 0 {
			_, err = fmt.Fprint(f, "0\n")
		} else {
			_, err = fmt.Fprintf(f, "%s ", m)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func dimacsHeader(bad *Result) (int, int) {
	nv := 0
	nc := 0
	for _, m := range bad.Invariant {
		if m == 0 {
			nc++
			continue
		}
		mv := int(m.Var())
		if mv > nv {
			nv = mv
		}
	}
	return nv, nc
}

// Remove tries to remove the output info, which is a recursive
// directory removal.  It returns non-nil error if this directory
// removal fails.  Once remove has been called, Store will fail.
func (o *Output) Remove() error {
	return os.RemoveAll(o.root)
}

// Verify verifies all traces and invariants given to `o`,
// returning a slice of errors if there are any, and nil otherwise.
//
// Verify is time limited to 1 second. Since Verify requires sat
// calls for invariant verification, one possibgle error is
// ErrTimeout.  Use TryVerify to specify a different time limit.
func (o *Output) Verify() []error {
	return o.TryVerify(time.Second)
}

// TryVerify tries to verify all traces and invariants given
// to `o`, each within `dur` time.
func (o *Output) TryVerify(dur time.Duration) []error {
	var res []error
	for i := range o.bads {
		errs := o.TryVerifyResult(i, dur)
		if errs != nil {
			res = append(res, errs...)
		}
	}
	return res
}

// TryVerifyResult
func (o *Output) TryVerifyResult(i int, dur time.Duration) []error {
	o.deadline = time.Now().Add(dur)
	return o.VerifyResult(i)
}

// VerifyResult verifies the results associated with Result
// at index i in the backing Result slice.
//
// It should only be called when the corresponding bad
// has either an invariant or trace associated with it.
func (o *Output) VerifyResult(i int) []error {
	_, terr := os.Stat(o.TracePath(i))
	_, ierr := os.Stat(o.InvariantPath(i))
	if os.IsNotExist(terr) && os.IsNotExist(ierr) {
		return []error{fmt.Errorf("nothing to verify")}
	}
	if os.IsNotExist(terr) {
		if ierr != nil {
			return []error{ierr}
		}
		return o.verifyInv(i)
	}
	if terr != nil {
		return []error{terr}
	}
	return o.verifyTrace(i)
}

type dimacsVis struct {
	sat *gini.Gini
	ms  []z.Lit
}

func (d *dimacsVis) Add(m z.Lit) {
	d.sat.Add(m)
	d.ms = append(d.ms, m)
}
func (d *dimacsVis) Init(v, c int) {
}
func (d *dimacsVis) Eof() {
}
func (o *Output) readInv(i int, g *gini.Gini) error {
	f, e := os.Open(o.InvariantPath(i))
	if e != nil {
		return e
	}
	defer f.Close()
	vis := &dimacsVis{sat: g}
	if err := dimacs.ReadCnf(f, vis); err != nil {
		return err
	}
	o.bads[i].Invariant = vis.ms
	return nil
}

func (o *Output) verifyInv(i int) []error {
	aig, err := o.Aiger()
	if err != nil {
		return []error{err}
	}
	sat := gini.New()
	trans := aig.Sys()
	var errors = o.verifyInvInit(trans, sat, o.bads[i])
	cms := make([]z.Lit, 0, 16)
	pri := NewPrimer(aig.Sys(), o.bads[i].M)
	trans.ToCnf(sat)
	// add invariant constraint
	if err := o.readInv(i, sat); err != nil {
		return []error{err}
	}
	inv := o.bads[i].Invariant
	for _, m := range inv {
		sat.Add(m)
	}
	// check inv -> not(bad) as inv and bad is unsat.
	sat.Assume(o.bads[i].M)
	switch sat.Try(time.Until(o.deadline)) {
	case 0:
		errors = append(errors, fmt.Errorf("ErrTimeout"))
		return errors
	case 1:
		errors = append(errors,
			fmt.Errorf("ErrInvImpliesNotBad: %s\n", o.bads[i].M))

	}
	// consecution
	for _, m := range inv {
		if m == 0 {
			res := sat.Try(time.Until(o.deadline))
			switch res {
			case -1:
			case 0:
				errors = append(errors, fmt.Errorf("ErrTimeout"))
			case 1:
				errors = append(errors,
					fmt.Errorf("ErrNonConsecutiveInvariant: %v", cms))
			}
			cms = cms[:0]
			continue
		}
		cms = append(cms, m)
		mp := pri.Prime(m)
		sat.Assume(mp.Not())
	}
	return errors
}

func (o *Output) verifyInvInit(trans *logic.S, sat *gini.Gini, bad *Result) []error {
	inv := bad.Invariant
	// initial condition
	maxV := z.Var(0)
	for _, m := range trans.Latches {
		switch trans.Init(m) {
		case trans.T:
			sat.Assume(m)
		case trans.F:
			sat.Assume(m.Not())
		}
		if m.Var() > maxV {
			maxV = m.Var()
		}
	}
	sat.Test(nil)
	defer sat.Untest()
	var res []error
	cs := make([]z.Lit, 0, 13)
	for _, m := range inv {
		if m == 0 {
			st := sat.Try(time.Until(o.deadline))
			switch st {
			case 0:
				res = append(res, fmt.Errorf("ErrTimeout"))
				return res
			case 1:
				res = append(res, fmt.Errorf("ErrInvariantInit[%d]: %v", bad.M, cs))
			}
			cs = cs[:0]
			continue
		}
		if m.Var() > maxV {
			continue
		}
		sat.Assume(m.Not())
	}
	if len(cs) != 0 {
		res = append(res, fmt.Errorf("ErrInvNotNullTerminated[%d]", bad.M))
	}
	return res
}

func (o *Output) verifyTrace(i int) []error {
	tr, err := o.Trace(i)
	if err != nil {
		return []error{err}
	}
	g, err := o.Aiger()
	if err != nil {
		return []error{err}
	}
	return tr.Verify(g.Sys())
}

// Trace tries to parse and return the trace associated with `i`th
// bad state info.  Trace does not cache.
func (o *Output) Trace(i int) (*Trace, error) {
	f, err := os.Open(o.TracePath(i))
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tr, err := DecodeTrace(f)
	if err != nil {
		return nil, err
	}
	return tr, nil
}

// Invariant places the invariant associated with bad state i in
// dst.
func (o *Output) Invariant(dst inter.Adder, i int) error {
	f, e := os.Open(o.InvariantPath(i))
	if e != nil {
		return e
	}
	defer f.Close()
	return nil
}

// Aiger tries to return the aiger for this problem
func (o *Output) Aiger() (*aiger.T, error) {
	p := o.AigerPath()
	p, e := filepath.EvalSymlinks(p)
	if e != nil {
		return nil, e
	}
	f, e := os.Open(p)
	if e != nil {
		return nil, e
	}
	defer f.Close()
	g, e := aiger.ReadBinary(f)
	if e != nil {
		return nil, e
	}
	return g, nil
}

// RootDir gives the root of the output directory.
func (o *Output) RootDir() string {
	return o.root
}

// AigerPath gives the path to the linked aiger.
func (o *Output) AigerPath() string {
	return filepath.Join(o.root, aigName)
}

// TracePath gives the path to trace associated with bad state i.
func (o *Output) TracePath(i int) string {
	return filepath.Join(o.root,
		fmt.Sprintf("%d%s", o.bads[i].M, traceExt))
}

// InvariantPath gives the path to the invariant associated with bad state i.
func (o *Output) InvariantPath(i int) string {
	return filepath.Join(o.root,
		fmt.Sprintf("%d%s", o.bads[i].M, invExt))
}

// ResultPath gives the path associated with storing Result meta-data,
// in json and parseable by json.Unmarshall.
func (o *Output) ResultPath(i int) string {
	return filepath.Join(o.root,
		fmt.Sprintf("%d%s", o.bads[i].M, badExt))
}
