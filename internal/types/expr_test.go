package types //nolint: revive,nolintlint

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestExprExist(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		filename string
		content  []byte
		expected bool
	}{
		{
			name:     "empty filename",
			filename: "",
			expected: false,
		},
		{
			name:     "non-existing file",
			filename: "none",
			expected: false,
		},
		{
			name:     "empty file",
			filename: "empty",
			content:  []byte(""),
			expected: false,
		},
		{
			name:     "non-empty file",
			filename: "file",
			content:  []byte("content"),
			expected: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var inputFilePath string

			if tc.filename != "" {
				inputFilePath = filepath.Join(t.TempDir(), tc.filename)

				if tc.content != nil {
					err := os.WriteFile(inputFilePath, tc.content, 0600)
					if err != nil {
						t.Fatal("Can't create JSON input file:", err)
					}
				}
			}

			result, err := ExprExist(inputFilePath)
			if err != nil {
				t.Fatal("Unexpected error:", err)
			}

			if result != tc.expected {
				t.Fatalf("Unexpected result: want %t, got %t", tc.expected, result)
			}
		})
	}
}

func TestExprJQ(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		jsonInput   string
		jqExpr      string
		expected    any
		expectedErr string
	}{
		{
			name:        "empty input file",
			jsonInput:   "",
			jqExpr:      ".",
			expectedErr: "json: EOF",
		},
		{
			name:        "invalid JSON input",
			jsonInput:   `{foo: bar}`,
			jqExpr:      ".",
			expectedErr: "json: invalid character 'f' looking for beginning of object key string",
		},
		{
			name:        "invalid jq expression",
			jsonInput:   `{"foo": "bar"}`,
			jqExpr:      ".[",
			expectedErr: "parsing jq expression: unexpected EOF",
		},
		{
			name:      "string from object",
			jsonInput: `{"foo": "bar"}`,
			jqExpr:    ".foo",
			expected:  "bar",
		},
		{
			name:      "object keys",
			jsonInput: `{"a": 1, "b": 2, "c": 3}`,
			jqExpr:    ". | keys",
			expected:  []any{"a", "b", "c"},
		},
		{
			name:      "object from array",
			jsonInput: `["a", "b", "c"]`,
			jqExpr:    ". | to_entries | map({(.value): (.key + 1)}) | add",
			expected:  map[string]any{"a": 1, "b": 2, "c": 3},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			inputFilePath := filepath.Join(t.TempDir(), "input.json")

			err := os.WriteFile(inputFilePath, []byte(tc.jsonInput), 0600)
			if err != nil {
				t.Fatal("Can't create JSON input file:", err)
			}

			result, err := ExprJQ(t.Context(), inputFilePath, tc.jqExpr)
			if err != nil {
				if err.Error() != tc.expectedErr {
					t.Fatalf("Unexpected error: want %q, got %q", tc.expectedErr, err)
				}

				return
			} else if tc.expectedErr != "" {
				t.Fatalf("Expected error %q, got none", tc.expectedErr)
			}

			if diff := cmp.Diff(tc.expected, result); diff != "" {
				t.Fatalf("Unexpected result (-want +got):\n%s", diff)
			}
		})
	}
}

func TestExprLoadJSON(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		filename    string
		jsonContent string
		expected    any
		expectedErr string
	}{
		{
			name:     "empty filename",
			filename: "",
			expected: nil,
		},
		{
			name:        "empty input file",
			filename:    "empty.json",
			jsonContent: "",
			expectedErr: "EOF",
		},
		{
			name:        "invalid JSON content",
			filename:    "invalid.json",
			jsonContent: `{foo: bar}`,
			expectedErr: "invalid character 'f' looking for beginning of object key string",
		},
		{
			name:        "JSON object",
			filename:    "object.json",
			jsonContent: `{"foo": "bar"}`,
			expected:    map[string]any{"foo": "bar"},
		},
		{
			name:        "JSON array",
			filename:    "array.json",
			jsonContent: `["a", "b", "c"]`,
			expected:    []any{"a", "b", "c"},
		},
		{
			name:        "null",
			filename:    "null.json",
			jsonContent: `null`,
			expected:    nil,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var inputFilePath string

			if tc.filename != "" {
				inputFilePath = filepath.Join(t.TempDir(), tc.filename)

				err := os.WriteFile(inputFilePath, []byte(tc.jsonContent), 0600)
				if err != nil {
					t.Fatal("Can't create JSON file:", err)
				}
			}

			result, err := ExprLoadJSON(inputFilePath)
			if err != nil {
				if err.Error() != tc.expectedErr {
					t.Fatalf("Unexpected error: want %q, got %q", tc.expectedErr, err)
				}

				return
			} else if tc.expectedErr != "" {
				t.Fatalf("Expected error %q, got none", tc.expectedErr)
			}

			if diff := cmp.Diff(tc.expected, result); diff != "" {
				t.Fatalf("Unexpected result (-want +got):\n%s", diff)
			}
		})
	}
}

func TestExprMerge(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		o1, o2   map[string]any
		expected any
	}{
		{
			name:     "first nil",
			o1:       nil,
			o2:       map[string]any{},
			expected: map[string]any{},
		},
		{
			name:     "last nil",
			o1:       map[string]any{},
			o2:       nil,
			expected: map[string]any{},
		},
		{
			name:     "both nil",
			o1:       nil,
			o2:       nil,
			expected: map[string]any{},
		},
		{
			name:     "different",
			o1:       map[string]any{"a": 1},
			o2:       map[string]any{"b": 2},
			expected: map[string]any{"a": 1, "b": 2},
		},
		{
			name:     "override",
			o1:       map[string]any{"a": 1},
			o2:       map[string]any{"a": 2},
			expected: map[string]any{"a": 2},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result := ExprMerge(tc.o1, tc.o2)
			if diff := cmp.Diff(tc.expected, result); diff != "" {
				t.Fatalf("Unexpected result (-want +got):\n%s", diff)
			}
		})
	}
}

func TestExprTitle(t *testing.T) {
	t.Parallel()

	cases := []struct {
		input    string
		expected string
	}{
		{
			input:    "",
			expected: "",
		},
		{
			input:    "foo",
			expected: "Foo",
		},
		{
			input:    "foo_bar",
			expected: "Foo bar",
		},
		{
			input:    "FOO",
			expected: "Foo",
		},
	}

	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			t.Parallel()

			result := ExprTitle(tc.input)
			if result != tc.expected {
				t.Fatalf("Unexpected result: want %q, got %q", tc.expected, result)
			}
		})
	}
}

func TestExprXPath(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		xmlInput    string
		xpathExpr   string
		expected    any
		expectedErr string
	}{
		{
			name:        "empty input file",
			xmlInput:    "",
			xpathExpr:   "/",
			expectedErr: "xmlquery: invalid XML document",
		},
		{
			name:        "invalid XML input",
			xmlInput:    `<foo>`,
			xpathExpr:   ".",
			expectedErr: "XML syntax error on line 1: unexpected EOF",
		},
		{
			name:        "invalid xpath expression",
			xmlInput:    `<foo>bar</foo>`,
			xpathExpr:   "@",
			expectedErr: "expression must evaluate to a node-set",
		},
		{
			name:      "string from object",
			xmlInput:  `<foo>bar</foo>`,
			xpathExpr: "foo",
			expected:  "bar",
		},
		{
			name:      "first array element",
			xmlInput:  `<list><item>a</item><item>b</item><item>c</item></list>`,
			xpathExpr: "list//item",
			expected:  "a",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			inputFilePath := filepath.Join(t.TempDir(), "input.xml")

			err := os.WriteFile(inputFilePath, []byte(tc.xmlInput), 0600)
			if err != nil {
				t.Fatal("Can't create XML input file:", err)
			}

			result, err := ExprXPath(inputFilePath, tc.xpathExpr)
			if err != nil {
				if err.Error() != tc.expectedErr {
					t.Fatalf("Unexpected error: want %q, got %q", tc.expectedErr, err)
				}

				return
			} else if tc.expectedErr != "" {
				t.Fatalf("Expected error %q, got none", tc.expectedErr)
			}

			if diff := cmp.Diff(tc.expected, result); diff != "" {
				t.Fatalf("Unexpected result (-want +got):\n%s", diff)
			}
		})
	}
}
