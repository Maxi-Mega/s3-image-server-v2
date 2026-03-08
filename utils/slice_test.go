package utils //nolint: revive,nolintlint

import (
	"math"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

type TC interface {
	getName() string
	getOutput() any
	getExpectedOutput() any
}

type mapTC[I, O any] struct {
	name           string
	input          []I
	fn             func(I) O
	expectedOutput []O
}

func (tc mapTC[I, O]) getName() string {
	return tc.name
}

func (tc mapTC[I, O]) getOutput() any {
	return Map(tc.input, tc.fn)
}

func (tc mapTC[I, O]) getExpectedOutput() any {
	return tc.expectedOutput
}

type filterTC[T any] struct {
	name           string
	input          []T
	fn             func(T) bool
	expectedOutput []T
}

func (tc filterTC[T]) getName() string {
	return tc.name
}

func (tc filterTC[T]) getOutput() any {
	return Filter(tc.input, tc.fn)
}

func (tc filterTC[T]) getExpectedOutput() any {
	return tc.expectedOutput
}

func TestMap(t *testing.T) {
	t.Parallel()

	cases := []TC{
		mapTC[string, string]{
			name:           "Strings capitalization",
			input:          []string{"First", "second", "THIRD"},
			fn:             strings.ToUpper,
			expectedOutput: []string{"FIRST", "SECOND", "THIRD"},
		},
		mapTC[float64, float64]{
			name:           "Numbers rounding",
			input:          []float64{5.6568, -3.1, 0.7419382},
			fn:             math.Round,
			expectedOutput: []float64{6, -3, 1},
		},
		mapTC[string, int]{
			name:           "Strings lengths",
			input:          []string{"a", "bb", "ccc"},
			fn:             func(str string) int { return len(str) },
			expectedOutput: []int{1, 2, 3},
		},
		mapTC[any, any]{
			name:           "Nil slice",
			input:          nil,
			fn:             nil, // shouldn't be called
			expectedOutput: []any{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.getName(), func(t *testing.T) {
			t.Parallel()

			output := tc.getOutput()
			if diff := cmp.Diff(tc.getExpectedOutput(), output); diff != "" {
				t.Fatalf("Unexpected result:\n%s", diff)
			}
		})
	}
}

func TestFilter(t *testing.T) {
	t.Parallel()

	cases := []TC{
		filterTC[int]{
			name:           "Keep even numbers",
			input:          []int{1, 2, 3, 4, 5, 6},
			fn:             func(n int) bool { return n%2 == 0 },
			expectedOutput: []int{2, 4, 6},
		},
		filterTC[string]{
			name:           "No match",
			input:          []string{"alpha", "beta", "gamma"},
			fn:             func(s string) bool { return strings.HasPrefix(s, "z") },
			expectedOutput: []string{},
		},
		filterTC[string]{
			name:           "All match",
			input:          []string{"ab", "ac"},
			fn:             func(s string) bool { return strings.HasPrefix(s, "a") },
			expectedOutput: []string{"ab", "ac"},
		},
		filterTC[any]{
			name:           "Nil slice",
			input:          nil,
			fn:             nil, // shouldn't be called
			expectedOutput: []any{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.getName(), func(t *testing.T) {
			t.Parallel()

			output := tc.getOutput()
			if diff := cmp.Diff(tc.getExpectedOutput(), output); diff != "" {
				t.Fatalf("Unexpected result:\n%s", diff)
			}
		})
	}
}
