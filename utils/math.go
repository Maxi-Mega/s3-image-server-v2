package utils //nolint: revive,nolintlint

import "cmp"

// Min returns the minimum number among all the given values.
func Min[T cmp.Ordered](x T, y ...T) T {
	minValue := x

	for _, n := range y {
		if n < minValue {
			minValue = n
		}
	}

	return minValue
}
