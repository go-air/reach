// Copyright 2018 Scott Cotton. All rights reserved.  Use of this source
// code is governed by a license that can be found in the License file.

package reach

import (
	"fmt"
	"time"

	"github.com/go-air/gini/z"
)

// Result holds info about bad states.
type Result struct {
	M         z.Lit         // the defining literal in the associated logic.S
	Status    int           // 1=reachable -1=unreachable 0=unknown
	Depth     int           // The depth of the analysis (= length of trace or depth of unreachability)
	Dur       time.Duration // If a timeout was specified, then its duration.
	Trace     *Trace        `json:"-"` // A trace (optional even if Reachable is true)
	Invariant []z.Lit       `json:"-"` // invariant in cnf.
}

func (b *Result) String() string {
	return fmt.Sprintf("bad[%s]: status=%s depth=%d dur=%s",
		b.M, b.FormatStatus(), b.Depth, b.Dur)
}

// FormatStatus returns
func (b *Result) FormatStatus() string {
	switch b.Status {
	case -1:
		return "unreachable"
	case 1:
		return "reachable"
	case 0:
		return "unknown"
	default:
		panic("unreachable")
	}
}

// IsSolved returns whether or not the result represents a
// solution of a bad state.
func (b *Result) IsSolved() bool {
	return b.Status != 0
}

// IsReachable returns if a bad state is known
// to be reachable.
func (b *Result) IsReachable() bool {
	return b.Status == 1
}

// SetReachable makes b known to be reachable.
// The trace `t` is optional and will be recorded.
func (b *Result) SetReachable(t *Trace) {
	b.Status = 1
	b.Trace = t
}

// IsUnreachable returns whether `b` is known to be unreachable.
func (b *Result) IsUnreachable() bool {
	return b.Status == -1
}

// SetUnreachable makes `b` known to be unreachable.
func (b *Result) SetUnreachable() {
	b.Status = -1
	b.Trace = nil
}

// Add is for storing an invarant proving unreachability
// in memory.  It is optional.
func (b *Result) Add(m z.Lit) {
	b.Invariant = append(b.Invariant, m)
}
