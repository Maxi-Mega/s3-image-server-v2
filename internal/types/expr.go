package types //nolint: revive,nolintlint

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"os"
	"reflect"
	"regexp"
	"strings"
	"sync/atomic"
	"time"

	"github.com/Maxi-Mega/s3-image-server-v2/internal/logger"

	"github.com/antchfx/xmlquery"
	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/ast"
	"github.com/expr-lang/expr/vm"
	"github.com/itchyny/gojq"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

const (
	ExprProductBasePath = "productBasePath"
	ExprGeonames        = "geonames"
	ExprLocalization    = "localization"
	ExprProductInfo     = "productInfo"
)

const evalTimeout = 5 * time.Second

type DynamicInputFile struct {
	S3Path   string
	CacheKey string
	Date     time.Time
}

type ExprEnv struct {
	Ctx   context.Context //nolint: containedctx // lives only the duration of the expression evaluation
	Files map[string]DynamicInputFile
	Exprs map[string]*vm.Program
}

var ExprFunctions = []expr.Option{ //nolint: gochecknoglobals
	// Call another expression with the current context and returns its result.
	expr.Function(
		"_call",
		func(params ...any) (any, error) {
			t0 := time.Now()
			defer func() {
				logger.Tracef("[expr] _call(%q) took %s", params[0], time.Since(t0))
			}()

			res, err := ExprCall(params[0].(string), params[1].(ExprEnv)) //nolint: forcetypeassert // already validated
			return res, wrapErr("_call", err)
		},
		new(func(exprName string) (any, error)), // env param will be injected at compile time
		new(func(exprName string, env ExprEnv) (any, error)),
	),
	// Returns whether a file matched by the given file selector has been cached.
	expr.Function(
		"_exist",
		func(params ...any) (any, error) {
			t0 := time.Now()
			defer func() {
				logger.Tracef("[expr] _exist(%q) took %s", params[0], time.Since(t0))
			}()

			file, err := fileFromSelector(params[0], params[1])
			if err != nil {
				return false, wrapErr("_exist", err)
			}

			res, err := ExprExist(file.CacheKey)
			return res, wrapErr("_exist", err)
		},
		new(func(fileSelector string) (bool, error)),
		new(func(fileSelector string, env ExprEnv) (bool, error)),
	),
	// Returns the result of the given jq expression on the file matched by the given file selector.
	expr.Function(
		"_jq",
		func(params ...any) (any, error) {
			t0 := time.Now()
			defer func() {
				logger.Tracef("[expr] _jq(%q, ...) took %s", params[1], time.Since(t0))
			}()

			file, err := fileFromSelector(params[1], params[3])
			if err != nil {
				return nil, wrapErr("_jq", err)
			}

			res, err := ExprJQ(params[0].(context.Context), file.CacheKey, params[2].(string)) //nolint: forcetypeassert // already validated
			return res, wrapErr("_jq", err)
		},
		new(func(ctx context.Context, fileSelector string, filter string) (any, error)),
		new(func(ctx context.Context, fileSelector string, filter string, env ExprEnv) (any, error)),
	),
	// Loads the content of the file matched by the given file selector to a JSON object.
	expr.Function(
		"_loadJSON",
		func(params ...any) (any, error) {
			t0 := time.Now()
			defer func() {
				logger.Tracef("[expr] _loadJSON(%q) took %s", params[0], time.Since(t0))
			}()

			file, err := fileFromSelector(params[0], params[1])
			if err != nil {
				return nil, wrapErr("_loadJSON", err)
			}

			res, err := ExprLoadJSON(file.CacheKey)
			return res, wrapErr("_loadJSON", err)
		},
		new(func(fileSelector string) (any, error)),
		new(func(fileSelector string, env ExprEnv) (any, error)),
	),
	// Merges the two given maps/objects into a new one, the second one overwriting the first one.
	expr.Function(
		"_merge",
		func(params ...any) (any, error) {
			t0 := time.Now()
			defer func() {
				logger.Tracef("[expr] _merge(...) took %s", time.Since(t0))
			}()

			return ExprMerge(params[0].(map[string]any), params[1].(map[string]any)), nil //nolint: forcetypeassert // already validated
		},
		new(func(o1, o2 map[string]any) (map[string]any, error)),
	),
	// Replaces matches of the regex on str with the replacement string.
	expr.Function(
		"_replaceRegex",
		func(params ...any) (any, error) {
			t0 := time.Now()
			defer func() {
				logger.Tracef("[expr] _replaceRegex(%q, %s, %s) took %s", params[0], params[1], params[2], time.Since(t0))
			}()

			res, err := ExprReplaceRegex(params[0].(string), params[1].(string), params[2].(string)) //nolint: forcetypeassert // already validated
			return res, wrapErr("_replaceRegex", err)
		},
		new(func(str string, regex string, replacement string) (string, error)),
	),
	// Returns the S3 path of the file matched by the given file selector.
	expr.Function(
		"_s3Key",
		func(params ...any) (any, error) {
			t0 := time.Now()
			defer func() {
				logger.Tracef("[expr] _merge(...) took %s", time.Since(t0))
			}()

			file, err := fileFromSelector(params[0], params[1])
			if err != nil {
				return "", wrapErr("_s3Key", err)
			}

			return file.S3Path, nil
		},
		new(func(fileSelector string) (string, error)),
		new(func(fileSelector string, env ExprEnv) (string, error)),
	),
	// Returns the titled version of the given string (first letter uppercase, rest lowercase, spaces replaced by underscores).
	expr.Function(
		"_title",
		func(params ...any) (any, error) {
			t0 := time.Now()
			defer func() {
				logger.Tracef("[expr] _title(%q) took %s", params[0], time.Since(t0))
			}()

			return ExprTitle(params[0].(string)), nil //nolint: forcetypeassert // already validated
		},
	),
	// Returns the result of the given xpath expression on the file matched by the given file selector.
	expr.Function(
		"_xpath",
		func(params ...any) (any, error) {
			t0 := time.Now()
			defer func() {
				logger.Tracef("[expr] _xpath(%q, ...) took %s", params[0], time.Since(t0))
			}()

			file, err := fileFromSelector(params[0], params[1])
			if err != nil {
				return nil, wrapErr("_xpath", err)
			}

			res, err := ExprXPath(file.CacheKey, params[1].(string)) //nolint: forcetypeassert // already validated
			return res, wrapErr("_xpath", err)
		},
		new(func(fileSelector string, xpath string) (any, error)),
		new(func(fileSelector string, xpath string, env ExprEnv) (any, error)),
	),
}

type ExprTestCounterKey struct{}

var ExprTestingFunctions = []expr.Option{ //nolint: gochecknoglobals
	expr.Function(
		"__testCounter__",
		func(params ...any) (any, error) {
			ctx := params[0].(context.Context) //nolint: forcetypeassert // already validated

			counter, ok := ctx.Value(ExprTestCounterKey{}).(*atomic.Int64)
			if !ok {
				return nil, errors.New("__testCounter__: context value not found")
			}

			counter.Add(1)

			return nil, nil //nolint: nilnil
		},
		new(func(ctx context.Context) (any, error)),
	),
}

type ExprEnvInjector struct{}

func (ExprEnvInjector) Visit(node *ast.Node) {
	funcsWithEnv := map[string]bool{
		"_call":     true,
		"_exist":    true,
		"_jq":       true,
		"_loadJSON": true,
		"_s3Key":    true,
		"_xpath":    true,
	}

	if callNode, ok := (*node).(*ast.CallNode); ok {
		if callee, ok := callNode.Callee.(*ast.IdentifierNode); ok && funcsWithEnv[callee.Value] {
			arity := callee.Type().NumIn()
			if arity == 0 || callee.Type().In(arity-1) != reflect.TypeFor[ExprEnv]() {
				callNode.Arguments = append(callNode.Arguments, &ast.IdentifierNode{Value: "$env"})
			}
		}
	}
}

func fileFromSelector(selParam, envParam any) (DynamicInputFile, error) {
	selector := selParam.(string) //nolint: forcetypeassert // already validated
	env := envParam.(ExprEnv)     //nolint: forcetypeassert // already validated

	file, found := env.Files[selector]
	if found {
		return file, nil
	}

	return DynamicInputFile{}, fmt.Errorf("unknown file selector %q", selParam)
}

func wrapErr(fn string, err error) error {
	if err != nil {
		return fmt.Errorf("%s: %w", fn, err)
	}

	return nil
}

func ExprCall(exprName string, env ExprEnv) (any, error) {
	prgm, found := env.Exprs[exprName]
	if !found {
		return nil, fmt.Errorf("_call: expr %q not found", exprName)
	}

	output, err := expr.Run(prgm, env)
	if err != nil {
		return nil, fmt.Errorf("_call: expr %q: %w", exprName, err)
	}

	return output, nil
}

func ExprExist(filePath string) (bool, error) {
	if filePath == "" {
		return false, nil
	}

	stat, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}

		return false, err //nolint: wrapcheck // wrapped by caller
	}

	if stat.Size() == 0 {
		return false, nil
	}

	return true, nil
}

func ExprJQ(ctx context.Context, filePath string, jqExpression string) (any, error) {
	if filePath == "" {
		return nil, nil //nolint: nilnil
	}

	query, err := gojq.Parse(jqExpression)
	if err != nil {
		return nil, fmt.Errorf("parsing jq expression: %w", err)
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, err //nolint: wrapcheck // wrapped by caller
	}

	defer file.Close()

	var input any

	err = json.NewDecoder(file).Decode(&input)
	if err != nil {
		return nil, fmt.Errorf("json: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, evalTimeout)
	defer cancel()

	iter := query.RunWithContext(ctx, input)

	v, ok := iter.Next()
	if !ok {
		return nil, nil //nolint: nilnil
	}

	if err, ok := v.(error); ok {
		if haltErr := new(gojq.HaltError); errors.As(err, &haltErr) && haltErr.Value() == nil {
			return nil, nil //nolint: nilnil
		}

		return nil, err
	}

	return v, nil
}

func ExprLoadJSON(filePath string) (any, error) {
	if filePath == "" {
		return nil, nil //nolint: nilnil
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, err //nolint: wrapcheck // wrapped by caller
	}

	defer file.Close()

	var result any

	err = json.NewDecoder(file).Decode(&result)
	if err != nil {
		return nil, err //nolint: wrapcheck // wrapped by caller
	}

	return result, nil
}

func ExprMerge(o1, o2 map[string]any) any {
	if o1 == nil { // avoid returning a nil map
		o1 = make(map[string]any, len(o2))
	}

	maps.Insert(o1, maps.All(o2))

	return o1
}

func ExprReplaceRegex(str, regex, replacement string) (string, error) {
	re, err := regexp.Compile(regex)
	if err != nil {
		return "", err //nolint: wrapcheck // wrapped by caller
	}

	return re.ReplaceAllString(str, replacement), nil
}

func ExprTitle(value string) string {
	value = cases.Title(language.English).String(value)
	return strings.ReplaceAll(value, "_", " ")
}

func ExprXPath(filePath string, xpathExpression string) (res any, err error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err //nolint: wrapcheck // wrapped by caller
	}

	defer f.Close()

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%s", r)
		}
	}()

	doc, err := xmlquery.Parse(f)
	if err != nil {
		return nil, err //nolint: wrapcheck // wrapped by caller
	}

	return xmlquery.FindOne(doc, xpathExpression).InnerText(), nil
}
