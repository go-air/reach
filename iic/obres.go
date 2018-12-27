package iic

import "fmt"

type obRes int

const (
	// the proof obligation was blocked.
	obBlocked obRes = iota
	// the proof obligation was backwards extended
	obExtended

	// timed out.
	obTimeout
)

func (o obRes) String() string {
	switch o {
	case obBlocked:
		return fmt.Sprintf("<res-blocked>")
	case obExtended:
		return fmt.Sprintf("<res-extended>")
	case obTimeout:
		return fmt.Sprintf("<res-timeout>")
	default:
		panic("unreachable")
	}
}
