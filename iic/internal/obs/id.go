package obs

import "fmt"

type Id int

func (i Id) String() string {
	return fmt.Sprintf("o%d", i)
}
