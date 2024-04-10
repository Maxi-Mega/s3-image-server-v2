package utils

import (
	"regexp"
	"strconv"
	"testing"
	"time"
)

func TestCommonPrefix(t *testing.T) {
	t.Parallel()

	cases := []struct {
		input    []string
		expected string
	}{
		{
			input:    []string{"PRODUCT/1", "PRODUCT/2", "PRODUCT/3"},
			expected: "PRODUCT/",
		},
		{
			input:    []string{"PRODUCT/1", "PRODUCT/2", "PROD_"},
			expected: "PROD",
		},
		{
			input:    []string{"First", "Second", "Third"},
			expected: "",
		},
		{
			input:    []string{"PRODUCT/1"},
			expected: "PRODUCT/1",
		},
		{
			input:    []string{},
			expected: "",
		},
	}

	for i, tc := range cases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Parallel()

			output := CommonPrefix(tc.input...)
			if output != tc.expected {
				t.Fatalf("Unexpected result for CommonPrefix(%v): got %q, want %q", tc.input, output, tc.expected)
			}
		})
	}
}

func TestGetRegexMatchGroup(t *testing.T) {
	t.Parallel()

	cases := []struct {
		re             *regexp.Regexp
		str            string
		group          int
		expectedResult string
		expectedMatch  bool
	}{
		{
			re:             regexp.MustCompile(`^(?P<parent>.*/QOF[^/]*/\d{4}/\d{2}/\d{2}/[^/]*)/preview\.jpg$`),
			str:            "product/IMAGE/ORTHO_PS/QOF15/2021/08/21/IMG_398/preview.jpg",
			group:          1,
			expectedResult: "product/IMAGE/ORTHO_PS/QOF15/2021/08/21/IMG_398",
			expectedMatch:  true,
		},
		{
			re:            regexp.MustCompile(`^(?P<parent>.*/QOF[^/]*/\d{4}/\d{2}/\d{2}/[^/]*)/preview\.jpg$`),
			str:           "product/IMAGE/ORTHO_PS/QOF15/2021/08/21/IMG_398/preview.png",
			group:         1,
			expectedMatch: false,
		},
		{
			re:            regexp.MustCompile(`^(?P<parent>.*/QOF[^/]*/\d{4}/\d{2}/\d{2}/[^/]*)/preview\.jpg$`),
			str:           "product/IMAGE/ORTHO_PS/QOF15/2021/08/21/IMG_398/preview.jpg",
			group:         2,
			expectedMatch: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.str, func(t *testing.T) {
			t.Parallel()

			result, match := GetRegexMatchGroup(tc.re, tc.str, tc.group)
			if match != tc.expectedMatch {
				t.Fatalf("Expected match=%v, got %v", tc.expectedMatch, match)
			}

			if result != tc.expectedResult {
				t.Fatalf("Expected result %q, got %q", tc.expectedResult, result)
			}
		})
	}
}

func TestFormatDuration(t *testing.T) {
	t.Parallel()

	cases := map[time.Duration]string{
		2 * time.Hour:                "2h0m0s",
		6*time.Hour + 45*time.Minute: "6h45m0s",
		25 * time.Hour:               "1d1h0m0s",
		48 * time.Hour:               "2d0s",
	}

	for input, expectedOutput := range cases {
		output := FormatDuration(input)
		if output != expectedOutput {
			t.Fatalf("Unexpected result for FormatDuration(%v): got %q, want %q", input, output, expectedOutput)
		}
	}
}
