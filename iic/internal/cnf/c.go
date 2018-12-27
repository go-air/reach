package cnf

import (
	"fmt"

	"github.com/irifrance/gini/z"
	"github.com/irifrance/reach/iic/internal/lits"
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
