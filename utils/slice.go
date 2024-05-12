package utils

func Map[I, O any](s []I, fn func(I) O) []O {
	result := make([]O, len(s))

	for i, e := range s {
		result[i] = fn(e)
	}

	return result
}
