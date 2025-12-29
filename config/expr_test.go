package config

import (
	"maps"
	"slices"
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

	exprByFunc := map[string]string{
		"_call":         `_call("__dummyFn__")`,
		"_exist":        `_exist("jsonFile")`,
		"_jq":           `_jq("jsonFile", ".key")`,
		"_loadJSON":     `_loadJSON("jsonFile")`,
		"_merge":        `_merge({"a": 1}, {"b": 2})`,
		"_replaceRegex": `_replaceRegex("some/value", "/", "@")`,
		"_s3Key":        `_s3Key("jsonFile")`,
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
		"_jq":           "value",
		"_loadJSON":     map[string]any{"key": "value"},
		"_merge":        map[string]any{"a": 1, "b": 2},
		"_replaceRegex": "some@value",
		"_s3Key":        "path/to/file.json",
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
				S3Path:   "path/to/file.json",
				CacheKey: "./testdata/file.json",
				Date:     time.Now(),
			},
			"xmlFile": {
				S3Path:   "path/to/file.xml",
				CacheKey: "./testdata/file.xml",
				Date:     time.Now(),
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
