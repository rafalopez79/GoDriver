package util

//Min of 2 ints
func Min(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

//Max of 2 ints
func Max(a int, b int) int {
	if a < b {
		return b
	}
	return a
}
