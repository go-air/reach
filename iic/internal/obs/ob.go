package obs

import (
	"fmt"

	"github.com/go-air/gini/z"
	"github.com/go-air/reach/iic/internal/lits"
)

// Proof obligation.
type ob struct {
	parent      Id
	k           int
	distToBad   int
	ms          lits.Span
	initWitness z.Lit
	sig         uint64
	nKids       int
	chosen      bool
}

func (o *ob) String() string {
	return fmt.Sprintf("k=%d dist=%d root=%t", o.k, o.distToBad,
		o.parent == 0)
}
