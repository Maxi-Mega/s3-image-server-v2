package utils

import "cmp"

// Min returns the minimum number among all the given values.
func Min[T cmp.Ordered](x T, y ...T) T {
	min := x

	for _, n := range y {
		if n < min {
			min = n
		}
	}

	return min
}
