package sim

type luby struct {
	exp   uint
	turns uint
}

func newLuby() *luby {
	return &luby{exp: 0, turns: 0}
}

func (l *luby) Next() uint {
	res := uint(1 << l.exp)
	if res&l.turns == 0 {
		l.exp = 0
		l.turns++
	} else {
		l.exp++
	}
	return res
}
