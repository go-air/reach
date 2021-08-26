// Copyright 2018 The Reach Authors. All rights reserved.  Use of this source
// code is governed by a license that can be found in the License file.

package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-air/gini/logic/aiger"
	"github.com/go-air/gini/z"
)

func readAiger(fn string) (*aiger.T, error) {
	f, err := os.Open(fn)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	ext := filepath.Ext(fn)
	switch ext {
	case ".aig":
		return aiger.ReadBinary(f)
	case ".aag":
		return aiger.ReadAscii(f)
	}
	return nil, fmt.Errorf("unknown filename extension '%s'", ext)
}

func aigerBad(g *aiger.T) []z.Lit {
	if len(g.Bad) == 0 {
		return g.Outputs
	}
	return g.Bad
}
