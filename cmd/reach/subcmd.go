// Copyright 2018 Scott Cotton. All rights reserved.  Use of this source
// code is governed by a license that can be found in the License file.

package main

import "flag"

type subCmd struct {
	Name  string
	Flags *flag.FlagSet
	Run   func(*subCmd, []string)
	Init  func(*subCmd)
	Usage string
	Short string
	Long  string
}
