package config

import (
	"maps"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/Maxi-Mega/s3-image-server-v2/internal/types"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/conf"
	"github.com/expr-lang/expr/vm"
	"github.com/google/go-cmp/cmp"
)

func TestExprFunctions(t *testing.T) {
	t.Parallel()

	date := time.Date(2026, 2, 26, 11, 34, 0, 0, time.Local) //nolint: gosmopolitan

	exprByFunc := map[string]string{
		"_call":         `_call("__dummyFn__")`,
		"_exist":        `_exist("jsonFile")`,
		"_fileDate":     `_fileDate("jsonFile", "2006-01-02 15:04:05")`,
		"_jq":           `_jq("jsonFile", ".key")`,
		"_loadJSON":     `_loadJSON("jsonFile")`,
		"_merge":        `_merge({"a": 1}, {"b": 2})`,
		"_replaceRegex": `_replaceRegex("some/value", "/", "@")`,
		"_s3Key":        `_s3Key("jsonFile")`,
		"_s3Uri":        `_s3Uri("jsonFile")`,
		"_title":        `_title("some str")`,
		"_xpath":        `_xpath("xmlFile", "//node")`,
	}

	exprCfg := conf.CreateNew()

	for _, opt := range types.ExprFunctions {
		opt(exprCfg)
	}

	expectedFuncNames := slices.Collect(maps.Keys(exprCfg.Functions))
	exprFuncNames := slices.Collect(maps.Keys(exprByFunc))

	slices.Sort(expectedFuncNames)
	slices.Sort(exprFuncNames)

	if diff := cmp.Diff(expectedFuncNames, exprFuncNames); diff != "" {
		t.Errorf("Tested expr functions don't match declared ones (-want +got):\n%s", diff)
	}

	expressions, err := parseExpressions(exprByFunc)
	if err != nil {
		t.Fatal(err)
	}

	expectedOutputs := map[string]any{
		"_call":         7,
		"_exist":        true,
		"_fileDate":     "2026-02-26 11:34:00",
		"_jq":           "value",
		"_loadJSON":     map[string]any{"key": "value"},
		"_merge":        map[string]any{"a": 1, "b": 2},
		"_replaceRegex": "some@value",
		"_s3Key":        "path/to/file.json",
		"_s3Uri":        "s3://bkt/path/to/file.json",
		"_title":        "Some Str",
		"_xpath":        "data",
	}

	dummyExpr, err := expr.Compile(`7`, expr.Env(types.ExprEnv{}))
	if err != nil {
		t.Fatal("Failed to compile dummy expr:", err)
	}

	env := types.ExprEnv{
		Ctx: t.Context(),
		Files: map[string]types.DynamicInputFile{
			"jsonFile": {
				S3Bucket: "bkt",
				S3Path:   "path/to/file.json",
				CacheKey: "./testdata/file.json",
				Date:     date,
			},
			"xmlFile": {
				S3Bucket: "bkt",
				S3Path:   "path/to/file.xml",
				CacheKey: "./testdata/file.xml",
				Date:     date,
			},
		},
		Exprs: map[string]*vm.Program{
			"__dummyFn__": dummyExpr,
		},
	}

	for name, prgm := range expressions {
		output, err := expr.Run(prgm, env)
		if err != nil {
			t.Fatalf("expr %q: %v", name, err)
		}

		if diff := cmp.Diff(expectedOutputs[name], output); diff != "" {
			t.Errorf("expr %q: unexpected output (-want +got):\n%s", name, diff)
		}
	}
}

func TestExprFunctionAdditionalBranches(t *testing.T) {
	t.Parallel()

	date := time.Date(2026, 2, 26, 11, 34, 0, 0, time.UTC)

	expressions, err := parseExpressions(map[string]string{
		"fileDateDefault": `_fileDate("jsonFile")`,
	})
	if err != nil {
		t.Fatal(err)
	}

	env := types.ExprEnv{
		Ctx: t.Context(),
		Files: map[string]types.DynamicInputFile{
			"jsonFile": {
				S3Bucket: "bkt",
				S3Path:   "path/to/file.json",
				CacheKey: "./testdata/file.json",
				Date:     date,
			},
		},
	}

	output, err := expr.Run(expressions["fileDateDefault"], env)
	if err != nil {
		t.Fatal(err)
	}

	if output != date.String() {
		t.Fatalf("Unexpected _fileDate default output: want %q, got %q.", date.String(), output)
	}
}

func TestExprFunctionErrors(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name          string
		expression    string
		files         map[string]types.DynamicInputFile
		exprs         map[string]*vm.Program
		expectedError string
	}{
		{
			name:          "unknown selector through s3 key",
			expression:    `_s3Key("missing")`,
			expectedError: `_s3Key: unknown file selector "missing"`,
		},
		{
			name:          "unknown selector through exist",
			expression:    `_exist("missing")`,
			expectedError: `_exist: unknown file selector "missing"`,
		},
		{
			name:          "unknown selector through file date",
			expression:    `_fileDate("missing")`,
			expectedError: `_fileDate: unknown file selector "missing"`,
		},
		{
			name:          "unknown selector through jq",
			expression:    `_jq("missing", ".")`,
			expectedError: `_jq: unknown file selector "missing"`,
		},
		{
			name:          "unknown selector through load json",
			expression:    `_loadJSON("missing")`,
			expectedError: `_loadJSON: unknown file selector "missing"`,
		},
		{
			name:          "unknown selector through s3 uri",
			expression:    `_s3Uri("missing")`,
			expectedError: `_s3Uri: unknown file selector "missing"`,
		},
		{
			name:          "unknown selector through xpath",
			expression:    `_xpath("missing", "//node")`,
			expectedError: `_xpath: unknown file selector "missing"`,
		},
		{
			name:          "missing called expression",
			expression:    `_call("missing")`,
			expectedError: `_call: expr "missing" not found`,
		},
		{
			name:       "called expression runtime failure",
			expression: `_call("bad")`,
			exprs: map[string]*vm.Program{
				"bad": mustParseExpression(t, `_s3Key("missing")`),
			},
			expectedError: `_call: expr "bad": _s3Key: unknown file selector "missing"`,
		},
		{
			name:          "jq error keeps wrapper prefix",
			expression:    `_jq("jsonFile", ".[")`,
			files:         jsonFileEnvFiles(),
			expectedError: `_jq: parsing jq expression`,
		},
		{
			name:          "test counter missing context value",
			expression:    `__testCounter__()`,
			expectedError: `__testCounter__: context value not found`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			expressions, err := parseExpressions(map[string]string{"expr": tc.expression})
			if err != nil {
				t.Fatal(err)
			}

			env := types.ExprEnv{
				Ctx:   t.Context(),
				Files: tc.files,
				Exprs: tc.exprs,
			}

			_, err = expr.Run(expressions["expr"], env)
			if err == nil {
				t.Fatal("Expected an error, got none.")
			}

			if !strings.Contains(err.Error(), tc.expectedError) {
				t.Fatalf("Expected error to contain %q, got %q.", tc.expectedError, err.Error())
			}
		})
	}
}

func mustParseExpression(t *testing.T, rawExpr string) *vm.Program {
	t.Helper()

	expressions, err := parseExpressions(map[string]string{"expr": rawExpr})
	if err != nil {
		t.Fatal(err)
	}

	return expressions["expr"]
}

func jsonFileEnvFiles() map[string]types.DynamicInputFile {
	return map[string]types.DynamicInputFile{
		"jsonFile": {
			S3Bucket: "bkt",
			S3Path:   "path/to/file.json",
			CacheKey: "./testdata/file.json",
		},
	}
}
