package utils //nolint: revive,nolintlint

import (
	"strconv"
	"testing"
)

func TestMin(t *testing.T) {
	t.Parallel()

	cases := []struct {
		getOutput func() any
		expected  any // must implement cmp.Ordered
	}{
		{
			getOutput: func() any { return Min(5, 7, -2, 3.14) },
			expected:  -2., // Min() will return a float
		},
		{
			getOutput: func() any { return Min(667) },
			expected:  667,
		},
		{
			getOutput: func() any { return Min(-0.654, -0.003, -0.078, -0.7002) },
			expected:  -0.7002,
		},
	}

	for i, tc := range cases {
		t.Run(strconv.Itoa(i+1), func(t *testing.T) {
			t.Parallel()

			output := tc.getOutput()
			if output != tc.expected {
				t.Fatalf("got %v, want %v", output, tc.expected)
			}
		})
	}
}
