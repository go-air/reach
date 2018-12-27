// Copyright 2018 Scott Cotton. All rights reserved.  Use of this source
// code is governed by a license that can be found in the License file.

package iic

import (
	"fmt"
	"io"
	"time"

	"github.com/irifrance/gini"
	"github.com/irifrance/gini/z"
)

// TBD: add relative calls and mean/stddev online duration.

// satmon is a sat wrapper that monitors time spent in sat calls and total
// number of sat calls.
type satmon struct {
	name     string
	calls    int64
	nSat     int64
	nUnsat   int64
	sat      *gini.Gini
	dur      time.Duration
	deadline *time.Time
}

func newSatMon(name string, sat *gini.Gini, deadline *time.Time) *satmon {
	return &satmon{name: name, sat: sat, deadline: deadline}
}

func (m *satmon) Try() int {
	start := time.Now()
	dur := time.Until(*m.deadline)
	res := m.sat.Try(dur)
	m.calls++
	switch res {
	case 1:
		m.nSat++
	case -1:
		m.nUnsat++
	}
	m.dur += time.Since(start)
	return res
}

func (m *satmon) Assume(ms ...z.Lit) {
	m.sat.Assume(ms...)
}

func (m *satmon) Why(ms []z.Lit) []z.Lit {
	return m.sat.Why(ms)
}

func (m *satmon) Test(d []z.Lit) (int, []z.Lit) {
	return m.sat.Test(d)
}

func (m *satmon) Untest() int {
	return m.sat.Untest()
}

func (m *satmon) Stats(w io.Writer) {
	fmt.Fprintf(w, "%s: %d sat calls in %s [%d/%d/%d]\n", m.name, m.calls,
		m.dur, m.nSat, m.nUnsat, m.calls-(m.nSat+m.nUnsat))
}

func (m *satmon) Value(n z.Lit) bool {
	return m.sat.Value(n)
}
