package main

func SliceInt64_isIn(sliceData []int64, elem int64) bool {
	for _, e := range sliceData {
		if e == elem {
			return true
		}
	}
	return false
}
