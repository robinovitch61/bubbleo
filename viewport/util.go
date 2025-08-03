package viewport

func clampValMinMax(v, minimum, maximum int) int {
	return max(minimum, min(maximum, v))
}
