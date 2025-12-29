package utils //nolint: revive,nolintlint

import "slices"

func Map[I, O any](s []I, fn func(I) O) []O {
	result := make([]O, len(s))

	for i, e := range s {
		result[i] = fn(e)
	}

	return result
}

func Filter[T any](s []T, fn func(T) bool) []T {
	result := make([]T, 0, len(s))

	for _, e := range s {
		if fn(e) {
			result = append(result, e)
		}
	}

	return slices.Clip(result)
}
