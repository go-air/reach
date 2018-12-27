// Copyright 2018 Scott Cotton. All rights reserved.  Use of this source
// code is governed by a license that can be found in the License file.

// Package sim provides simulation capabilities.
//
// sim.T is a simulator which implicitly parallelizes Boolean operations
// by doing 64 independent simulation steps per logic gate in one
// 64 bit word bitwise operation.
//
// Interfaces are provided for watches and monitoring.
package sim
