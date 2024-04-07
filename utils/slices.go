package utils

import "slices"

func RemoveFromSlice[T comparable](s []T, e T) []T {
	return slices.DeleteFunc(s, func(t T) bool {
		return t == e
	})
}
