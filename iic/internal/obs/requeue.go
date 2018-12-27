package obs

// Requeue is a function which returns true if an
// proof obligation should be kept in the queue.
//
// A minimal requirement is to return false if
// k >= max
var Requeue = RequeueLong

// Requeue function which allows for searching
// for deep counterexamples
func RequeueLong(k, d, max int) bool {
	return k < max
}

// Requeue function which only allows
// for searching for length max+1 counterexamples.
func RequeueShort(k, d, max int) bool {
	return k+d < max
}
