package utils //nolint: revive,nolintlint

import (
	"math"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

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

func TestMap(t *testing.T) {
	t.Parallel()

	cases := []interface {
		getName() string
		getOutput() any
		getExpectedOutput() any
	}{
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
