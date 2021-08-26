// Copyright (c) 2021 The Reach authors (see AUTHORS)
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of
// this software and associated documentation files (the "Software"), to deal in
// the Software without restriction, including without limitation the rights to
// use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
// the Software, and to permit persons to whom the Software is furnished to do so,
// subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
// FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
// COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
// CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package cnf

import (
	"fmt"

	"github.com/go-air/gini/z"
	"github.com/go-air/reach/iic/internal/lits"
)

type Id int

func (c Id) String() string {
	return fmt.Sprintf("c%d", c)
}

// clause represents a clause in T.
type clause struct {
	id    Id
	act   z.Lit
	ms    lits.Span
	sig   uint64
	level int
	rm    bool
}

// String implements Stringer.
func (c *clause) String() string {
	return fmt.Sprintf("c%d@%d (%p)", c.id, c.level, c)
}

// Id returns a uniq, packed integer id for `c`.
func (c *clause) Id() Id {
	return c.id
}

// ActLit returns the activation literal for `c`.
func (c *clause) ActLit() z.Lit {
	return c.act
}

func (c *clause) rmd() bool {
	return c.rm
}
