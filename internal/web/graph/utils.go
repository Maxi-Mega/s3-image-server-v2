package graph

import "errors"

var errInvalidTimeRange = errors.New("invalid time range")

func toMapStringAny[V any](m map[string]V) map[string]any {
	result := make(map[string]any, len(m))

	for k, v := range m {
		result[k] = v
	}

	return result
}
