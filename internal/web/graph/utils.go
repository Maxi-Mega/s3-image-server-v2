package graph

func toAnySlice[E any](s []E) []any {
	result := make([]any, len(s))

	for i, e := range s {
		result[i] = e
	}

	return result
}

func toMapStringAny[V any](m map[string]V) map[string]any {
	result := make(map[string]any, len(m))

	for k, v := range m {
		result[k] = v
	}

	return result
}
