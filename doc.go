// Copyright 2018 Scott Cotton. All rights reserved.  Use of this source
// code is governed by a license that can be found in the License file.

// Package reach contains finite state symbolic reachability libraries and
// tools.
//
// All reachability checking in this package and its subpackages uses a
// sequential circuit in the form of
// http://godoc.org/github.com/irifrance/gini/logic#S to represent a transition
// relation between state variables. State variables are latches in `S`. The
// set of initial states is defined by the initial values of latches in `S` and
// the bad states are defined by a literal in `S`.
//
//   The root package, reach, is centered around a few data structures to to
//   aid in coordinating checkers.
//
//   1. `Output`, which is the output of a checker from analyzing a logic.S and
//   a, or some, bad states.
//
//   2. `Result`, which is the result of analysing a single reachability query
//   with one set of bad states represented by a z.Lit in a logic.S.
//
//   3. `Primer` which can find `x'` given `x` for latches and properties.
//
//   4. `Trace`, which is a trace of a sequential logic system.
//
// These are concepts related to coordination of checkers. Various checkers are
// found in the subpackages of reach.
package reach
