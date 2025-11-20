package config

import (
	"errors"
	"fmt"
	"maps"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/Maxi-Mega/s3-image-server-v2/internal/types"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/ast"
	"github.com/expr-lang/expr/vm"
	"gopkg.in/yaml.v3"
)

const defaultCacheDirName = "s3_image_server"

var fullProductSignedURLRegexp = regexp.MustCompile(`fullProductSignedURL\((.*)\)`)

var (
	errInvalidConfig          = errors.New("the config is invalid")
	errNoImageGroupsSpecified = errors.New("no image groups specified")
	errDuplicate              = errors.New("duplicate")
	errTooHighValue           = errors.New("too high value")
)

func Load(configPath string) (Config, []string, error) {
	file, err := os.Open(configPath)
	if err != nil {
		return Config{}, nil, err //nolint:wrapcheck
	}

	defer file.Close()

	var cfg = defaultConfig()

	err = yaml.NewDecoder(file).Decode(&cfg)
	if err != nil {
		return Config{}, nil, fmt.Errorf("failed to parse config: %w", err)
	}

	warnings, err := cfg.validate()
	if err != nil {
		return Config{}, warnings, fmt.Errorf("%w: %w", errInvalidConfig, err)
	}

	err = cfg.process()
	if err != nil {
		return Config{}, warnings, fmt.Errorf("%w: %w", errInvalidConfig, err)
	}

	return cfg, warnings, nil
}

func (cfg *Config) validate() ([]string, error) {
	warnings := make([]string, 0)
	errs := make([]error, 0)

	dynamicFilterNames := make(map[string]bool, len(cfg.Products.DynamicFilters))

	for i, filter := range cfg.Products.DynamicFilters {
		if filter.Name == "" { //nolint: gocritic
			errs = append(errs, fmt.Errorf("empty name for dynamic filter n°%d", i+1))
		} else if dynamicFilterNames[filter.Name] {
			errs = append(errs, fmt.Errorf("duplicate dynamic filter name %q", filter.Name))
		} else {
			dynamicFilterNames[filter.Name] = true
		}

		if filter.Expression == "" {
			errs = append(errs, fmt.Errorf("empty expression for dynamic filter n°%d", i+1))
		}
	}

	err := validateFileSelectors(cfg.Products.DynamicData.FileSelectors)
	if err != nil {
		errs = append(errs, fmt.Errorf("invalid products file selectors: %w", err))
	}

	if len(cfg.Products.ImageGroups) == 0 {
		errs = append(errs, errNoImageGroupsSpecified)
	}

	imageGroupNames := make(map[string]bool)

	for _, grp := range cfg.Products.ImageGroups {
		if imageGroupNames[grp.GroupName] {
			errs = append(errs, fmt.Errorf("image group name %q is %w", grp.GroupName, errDuplicate))

			break
		}

		err := validateFileSelectors(grp.DynamicData.FileSelectors)
		if err != nil {
			errs = append(errs, fmt.Errorf("invalid file selectors in group %q: %w", grp.GroupName, err))
		}

		imageGroupNames[grp.GroupName] = true
		imageTypeNames := make(map[string]bool)

		for _, typ := range grp.Types {
			if imageTypeNames[typ.Name] {
				errs = append(errs, fmt.Errorf("image type name %q of group %q is %w", typ.Name, grp.GroupName, errDuplicate))

				break
			}

			for _, objType := range []string{types.ObjectPreview, types.ExprGeonames, types.ExprLocalization} {
				_, ok := typ.DynamicData.FileSelectors[objType]
				if !ok {
					_, ok = grp.DynamicData.FileSelectors[objType]
					if !ok {
						_, ok = cfg.Products.DynamicData.FileSelectors[objType]
						if !ok {
							warnings = append(warnings, fmt.Sprintf("no file selector provided for object type %q, in type %q/%q", objType, grp.GroupName, typ.Name))
						}
					}
				}
			}

			err := validateFileSelectors(typ.DynamicData.FileSelectors)
			if err != nil {
				errs = append(errs, fmt.Errorf("invalid file selectors in type %q/%q: %w", typ.Name, grp.GroupName, err))
			}

			imageTypeNames[typ.Name] = true
		}
	}

	if cfg.UI.ScaleInitialPercentage > math.MaxInt {
		errs = append(errs, fmt.Errorf("ui.scaleInitialPercentage has a %w (%d)", errTooHighValue, cfg.UI.ScaleInitialPercentage))
	}

	if cfg.UI.MaxImagesDisplayCount > math.MaxInt {
		errs = append(errs, fmt.Errorf("ui.maxImagesDisplayCount as a %w (%d)", errTooHighValue, cfg.UI.MaxImagesDisplayCount))
	}

	return warnings, errors.Join(errs...)
}

func validateFileSelectors(fileSelectors map[string]FileSelector) error {
	for name, selector := range fileSelectors {
		if selector.Kind != FileSelectorKindCached && selector.Kind != FileSelectorKindSignedURL && !fullProductSignedURLRegexp.MatchString(selector.Kind) {
			return fmt.Errorf("selector %q: unknown kind %q (accepted values are: %q, %q, %q)", name, selector.Kind, FileSelectorKindCached, FileSelectorKindSignedURL, FileSelectorKindFullProductSignedURL)
		}
	}

	return nil
}

func (cfg *Config) process() (err error) {
	if idx := strings.Index(cfg.S3.Endpoint, "://"); idx >= 0 {
		cfg.S3.Endpoint = cfg.S3.Endpoint[idx+3:]
	}

	cfg.Cache.CacheDir, err = filepath.Abs(cfg.Cache.CacheDir)
	if err != nil {
		return fmt.Errorf("could not resolve cache dir: %w", err)
	}

	cfg.Cache.CacheDir = filepath.Join(cfg.Cache.CacheDir, defaultCacheDirName)

	if cfg.UI.BaseURL == "" {
		cfg.UI.BaseURL = "/"
	}

	cfg.Products.TargetRelativeRgx, err = regexp.Compile(cfg.Products.TargetRelativeRegexp)
	if err != nil {
		return fmt.Errorf("can't parse products.targetRelativeRegexp: %w", err)
	}

	for g, imgGroup := range cfg.Products.ImageGroups {
		cfg.Products.ImageGroups[g].DynamicData = mergeDynamicData(imgGroup.DynamicData, cfg.Products.DynamicData)

		for t, imgType := range imgGroup.Types {
			cfg.Products.ImageGroups[g].Types[t].DynamicData = mergeDynamicData(imgType.DynamicData, cfg.Products.ImageGroups[g].DynamicData)

			err = ParseDynamicData(imgGroup.GroupName, imgType.Name, &cfg.Products.ImageGroups[g].Types[t].DynamicData)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func mergeDynamicData(child, parent DynamicData) DynamicData {
	result := DynamicData{
		FileSelectors: maps.Clone(parent.FileSelectors),
		Expressions:   maps.Clone(parent.Expressions),
	}

	if result.FileSelectors == nil {
		result.FileSelectors = make(map[string]FileSelector, len(child.FileSelectors))
	}

	if result.Expressions == nil {
		result.Expressions = make(map[string]string, len(child.Expressions))
	}

	maps.Insert(result.FileSelectors, maps.All(child.FileSelectors))
	maps.Insert(result.Expressions, maps.All(child.Expressions))

	return result
}

func ParseDynamicData(imgGroup, imgType string, dynData *DynamicData) error {
	err := parseFileSelectors(dynData.FileSelectors)
	if err != nil {
		return fmt.Errorf("can't parse products.imageGroups[%q].types[%q].dynamicData.fileSelectors: %w", imgGroup, imgType, err)
	}

	dynData.ExpressionsPrograms, err = parseExpressions(dynData.Expressions)
	if err != nil {
		return fmt.Errorf("can't parse products.imageGroups[%q].types[%q].dynamicData.expressions: %w", imgGroup, imgType, err)
	}

	for name, selector := range dynData.FileSelectors {
		switch selector.Kind {
		case FileSelectorKindSignedURL:
			selector.Link = true
		case FileSelectorKindFullProductSignedURL:
			_, found := dynData.Expressions[selector.KindParams[0]]
			if !found {
				return fmt.Errorf("invalid products.imageGroups[%q].types[%q].dynamicData.fileSelectors[%q]: fullProductSignedURL references the expression %q which is not defined", imgGroup, imgType, name, selector.KindParams[0])
			}

			selector.Link = true
		}
	}

	return nil
}

func parseFileSelectors(fileSelectors map[string]FileSelector) error {
	var err error

	for name, selector := range fileSelectors {
		selector.Rgx, err = regexp.Compile(selector.Regex)
		if err != nil {
			return fmt.Errorf("%q: %w", name, err)
		}

		matches := fullProductSignedURLRegexp.FindStringSubmatch(selector.Kind)
		if len(matches) == 2 && matches[1] != "" {
			selector.Kind = FileSelectorKindFullProductSignedURL
			selector.KindParams = strings.Split(strings.ReplaceAll(matches[1], " ", ""), ",")
		} else if matches != nil {
			return fmt.Errorf("%q: invalid fullProductSignedURL expression %q", name, selector.Kind)
		}

		fileSelectors[name] = selector
	}

	return nil
}

func parseExpressions(expressions map[string]string) (map[string]*vm.Program, error) {
	result := make(map[string]*vm.Program, len(expressions))

	options := append( //nolint: gocritic
		types.ExprFunctions,
		expr.Env(types.ExprEnv{}),
		expr.WithContext("Ctx"),
		expr.Patch(types.ExprEnvInjector{}),
	)

	if testing.Testing() {
		options = append(options, types.ExprTestingFunctions...)
	}

	for name, rawExpr := range expressions {
		program, err := expr.Compile(rawExpr, options...)
		if err != nil {
			return nil, fmt.Errorf("expression %q: %w", name, err)
		}

		err = validateExpression(program.Node())
		if err != nil {
			return nil, fmt.Errorf("expression %q: %w", name, err)
		}

		result[name] = program
	}

	return result, nil
}

type exprValidator struct {
	errs []error
}

func (ev *exprValidator) Visit(node *ast.Node) {
	if callNode, ok := (*node).(*ast.CallNode); ok { //nolint: nestif
		if callee, ok := callNode.Callee.(*ast.IdentifierNode); ok {
			switch callee.Value { //nolint: gocritic
			case "_replaceRegex":
				regexParam, ok := callNode.Arguments[1].(*ast.StringNode)
				if !ok {
					ev.errs = append(ev.errs, fmt.Errorf("_replaceRegex: second argument must be a string, not %s", callNode.Arguments[1].Nature()))
					return
				}

				if _, err := regexp.Compile(regexParam.Value); err != nil {
					ev.errs = append(ev.errs, fmt.Errorf("_replaceRegex: %w", err))
					return
				}
			}
		}
	}
}

func validateExpression(node ast.Node) error {
	var v exprValidator

	ast.Walk(&node, &v)

	return errors.Join(v.errs...)
}
