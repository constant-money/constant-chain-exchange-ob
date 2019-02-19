package orderbook

func Float64ComparatorAsc(a, b interface{}) int {
	aAsserted := a.(float64)
	bAsserted := b.(float64)
	switch {
	case aAsserted > bAsserted:
		return 1
	case aAsserted < bAsserted:
		return -1
	default:
		return 0
	}
}

func Float64ComparatorDesc(a, b interface{}) int {
	aAsserted := a.(float64)
	bAsserted := b.(float64)
	switch {
	case aAsserted < bAsserted:
		return 1
	case aAsserted > bAsserted:
		return -1
	default:
		return 0
	}
}
