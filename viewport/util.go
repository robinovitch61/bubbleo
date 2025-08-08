package viewport

func clampValZeroToMax(v, maximum int) int {
	return max(0, min(maximum, v))
}
