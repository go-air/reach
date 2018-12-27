package obs

type obq Set

func (o *obq) Len() int {
	return len(o.ks[o.k])
}

func (o *obq) Swap(i, j int) {
	sl := o.ks[o.k]
	sl[i], sl[j] = sl[j], sl[i]
}

func (o *obq) Less(i, j int) bool {
	sl := o.ks[o.k]
	return Less((*Set)(o), sl[i], sl[j])
}

func (o *obq) Push(x interface{}) {
	sl := o.ks[o.k]
	id := x.(Id)
	sl = append(sl, id)
	o.ks[o.k] = sl
}

func (o *obq) Pop() interface{} {
	sl := o.ks[o.k]
	n := len(sl) - 1
	res := sl[n]
	o.ks[o.k] = sl[:n]
	return res
}
