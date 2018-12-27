package obs

// Less defines how to prioritize proof obligations
// under the assumption they all are at the same level.
var Less = func(s *Set, a, b Id) bool {
	da, db := s.DistToBad(a), s.DistToBad(b)
	if da > db {
		return true
	}
	if da < db {
		return false
	}
	la, lb := len(s.Ms(a)), len(s.Ms(b))
	if la < lb {
		return true
	}
	if la > lb {
		return false
	}
	return a < b
}
