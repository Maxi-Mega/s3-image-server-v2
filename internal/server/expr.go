package server

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"maps"
	"path/filepath"
	"slices"
	"sync"
	"time"

	"github.com/Maxi-Mega/s3-image-server-v2/config"
	"github.com/Maxi-Mega/s3-image-server-v2/internal/logger"
	"github.com/Maxi-Mega/s3-image-server-v2/internal/s3"
	"github.com/Maxi-Mega/s3-image-server-v2/internal/types"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"
	"github.com/go-viper/mapstructure/v2"
)

var (
	errMissingExpression    = errors.New("missing expression")
	errUnexpectedOutputType = errors.New("unexpected output type")
)

const exprCacheTTL = 10 * time.Minute

type exprCacheKey struct {
	bucket   string
	s3key    string
	exprName string
}

type exprCacheEntry struct {
	sum   string
	ts    time.Time
	value any
}

type expressionManager struct {
	cacheDir   string
	dynFilters map[string]string
	// map[img group][img type][expr name] -> expr
	exprs map[string]map[string]map[string]*vm.Program
	// map[img group][img type] -> selectors
	fileSelectors map[string]map[string][]string
	l             sync.Mutex
	cacheSums     map[exprCacheKey]exprCacheEntry
}

func newExpressionManager(cfg config.Config) *expressionManager {
	dynamicFilters := make(map[string]string, len(cfg.Products.DynamicFilters))
	exprs := make(map[string]map[string]map[string]*vm.Program)
	selectors := make(map[string]map[string][]string)

	for _, filter := range cfg.Products.DynamicFilters {
		dynamicFilters[filter.Name] = filter.Expression
	}

	for _, group := range cfg.Products.ImageGroups {
		exprs[group.GroupName] = make(map[string]map[string]*vm.Program)
		selectors[group.GroupName] = make(map[string][]string)

		for _, imgType := range group.Types {
			exprs[group.GroupName][imgType.Name] = maps.Clone(imgType.DynamicData.ExpressionsPrograms)
			selectors[group.GroupName][imgType.Name] = slices.Collect(maps.Keys(imgType.DynamicData.FileSelectors))
		}
	}

	return &expressionManager{
		cacheDir:      cfg.Cache.CacheDir,
		dynFilters:    dynamicFilters,
		exprs:         exprs,
		fileSelectors: selectors,
		cacheSums:     make(map[exprCacheKey]exprCacheEntry),
	}
}

func (exprMan *expressionManager) getCache(imgBucket, imgKey, exprName string, sum string) (any, bool) {
	exprMan.l.Lock()
	defer exprMan.l.Unlock()

	entry, found := exprMan.cacheSums[exprCacheKey{imgBucket, imgKey, exprName}]
	if found && entry.sum == sum && time.Since(entry.ts) < exprCacheTTL {
		return entry.value, true
	}

	return nil, false
}

func (exprMan *expressionManager) updateCache(imgBucket, imgKey, exprName string, sum string, value any) {
	exprMan.l.Lock()
	defer exprMan.l.Unlock()

	exprMan.cacheSums[exprCacheKey{imgBucket, imgKey, exprName}] = exprCacheEntry{
		sum:   sum,
		ts:    time.Now(),
		value: value,
	}
}

func (exprMan *expressionManager) productBasePath(ctx context.Context, imgGroup, imgType string, s3event s3.Event) (string, error) {
	pbpExpr, found := exprMan.exprs[imgGroup][imgType][types.ExprProductBasePath]
	if !found {
		return "", fmt.Errorf("%w %q", errMissingExpression, types.ExprProductBasePath)
	}

	env := types.ExprEnv{
		Ctx: ctx,
		Files: map[string]types.DynamicInputFile{
			types.ObjectPreview: {
				S3Path: s3event.ObjectKey,
				Date:   s3event.ObjectLastModified,
			},
		},
		Exprs: exprMan.exprs[imgGroup][imgType],
	}
	selectorsSum := dynamicFilesChecksum(env.Files)

	if value, ok := exprMan.getCache(s3event.Bucket, s3event.ObjectKey, types.ExprProductBasePath, selectorsSum); ok {
		return value.(string), nil //nolint: forcetypeassert
	}

	output, err := expr.Run(pbpExpr, env)
	if err != nil {
		return "", fmt.Errorf("expr: %w", err)
	}

	basePath, ok := output.(string)
	if !ok {
		return "", fmt.Errorf("%w: want string, got %T", errUnexpectedOutputType, output)
	}

	exprMan.updateCache(s3event.Bucket, s3event.ObjectKey, types.ExprProductBasePath, selectorsSum, basePath)

	return basePath, nil
}

func (exprMan *expressionManager) imageGeonames(ctx context.Context, img image) (*types.Geonames, error) {
	geoExpr, found := exprMan.exprs[img.imgGroup][img.imgType][types.ExprGeonames]
	if !found {
		return nil, nil //nolint: nilnil
	}

	env := types.ExprEnv{
		Ctx:   ctx,
		Files: exprMan.valueMap2FilesMap(img),
		Exprs: exprMan.exprs[img.imgGroup][img.imgType],
	}
	selectorsSum := dynamicFilesChecksum(env.Files)

	if value, ok := exprMan.getCache(img.bucket, img.s3Key, types.ExprGeonames, selectorsSum); ok {
		return value.(*types.Geonames), nil //nolint: forcetypeassert
	}

	output, err := expr.Run(geoExpr, env)
	if err != nil {
		return nil, fmt.Errorf("expr: %w", err)
	}

	if output == nil {
		return nil, nil //nolint: nilnil
	}

	geonames := types.Geonames{
		CachedObject: types.CachedObject{
			LastModified: img.lastModified,
		},
	}

	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:  &geonames.Objects,
		TagName: "json",
	})
	if err != nil {
		return nil, err //nolint: wrapcheck
	}

	err = decoder.Decode(output)
	if err != nil {
		return nil, err //nolint: wrapcheck
	}

	exprMan.updateCache(img.bucket, img.s3Key, types.ExprGeonames, selectorsSum, &geonames)

	return &geonames, nil
}

func (exprMan *expressionManager) imageLocalization(ctx context.Context, img image) (*types.Localization, error) {
	locExpr, found := exprMan.exprs[img.imgGroup][img.imgType][types.ExprLocalization]
	if !found {
		return nil, nil //nolint: nilnil
	}

	env := types.ExprEnv{
		Ctx:   ctx,
		Files: exprMan.valueMap2FilesMap(img),
		Exprs: exprMan.exprs[img.imgGroup][img.imgType],
	}
	selectorsSum := dynamicFilesChecksum(env.Files)

	if value, ok := exprMan.getCache(img.bucket, img.s3Key, types.ExprLocalization, selectorsSum); ok {
		return value.(*types.Localization), nil //nolint: forcetypeassert
	}

	output, err := expr.Run(locExpr, env)
	if err != nil {
		return nil, fmt.Errorf("expr: %w", err)
	}

	if output == nil {
		return nil, nil //nolint: nilnil
	}

	localization := types.Localization{
		CachedObject: types.CachedObject{
			LastModified: img.lastModified,
		},
	}

	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:  &localization,
		TagName: "json",
	})
	if err != nil {
		return nil, err //nolint: wrapcheck
	}

	err = decoder.Decode(output)
	if err != nil {
		return nil, err //nolint: wrapcheck
	}

	exprMan.updateCache(img.bucket, img.s3Key, types.ExprLocalization, selectorsSum, &localization)

	return &localization, nil
}

func (exprMan *expressionManager) productInfo(ctx context.Context, img image) (*types.ProductInformation, error) {
	locExpr, found := exprMan.exprs[img.imgGroup][img.imgType][types.ExprProductInfo]
	if !found {
		return nil, nil //nolint: nilnil
	}

	env := types.ExprEnv{
		Ctx:   ctx,
		Files: exprMan.valueMap2FilesMap(img),
		Exprs: exprMan.exprs[img.imgGroup][img.imgType],
	}
	selectorsSum := dynamicFilesChecksum(env.Files)

	if value, ok := exprMan.getCache(img.bucket, img.s3Key, types.ExprProductInfo, selectorsSum); ok {
		return value.(*types.ProductInformation), nil //nolint: forcetypeassert
	}

	output, err := expr.Run(locExpr, env)
	if err != nil {
		return nil, fmt.Errorf("expr: %w", err)
	}

	if output == nil {
		return nil, nil //nolint: nilnil
	}

	var productInformation types.ProductInformation

	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:  &productInformation,
		TagName: "json",
	})
	if err != nil {
		return nil, err //nolint: wrapcheck
	}

	err = decoder.Decode(output)
	if err != nil {
		return nil, err //nolint: wrapcheck
	}

	exprMan.updateCache(img.bucket, img.s3Key, types.ExprProductInfo, selectorsSum, &productInformation)

	return &productInformation, nil
}

func (exprMan *expressionManager) dynamicFilters(ctx context.Context, img image) (map[string]string, error) {
	dynFilters := make(map[string]string, len(exprMan.dynFilters))

	env := types.ExprEnv{
		Ctx:   ctx,
		Files: exprMan.valueMap2FilesMap(img),
		Exprs: exprMan.exprs[img.imgGroup][img.imgType],
	}
	selectorsSum := dynamicFilesChecksum(env.Files)

	for name, exprName := range exprMan.dynFilters {
		if value, ok := exprMan.getCache(img.bucket, img.s3Key, exprName, selectorsSum); ok {
			valueStr, ok := value.(string)
			if ok {
				dynFilters[name] = valueStr
			} else {
				logger.Warnf("filter %q: expected a string result, got %T", name, value)

				dynFilters[name] = fmt.Sprint(value)
			}

			continue
		}

		prgm, ok := exprMan.exprs[img.imgGroup][img.imgType][exprName]
		if !ok {
			continue
		}

		output, err := expr.Run(prgm, env)
		if err != nil {
			return nil, fmt.Errorf("filter %q: %w", name, err)
		}

		dynFilters[name], ok = output.(string)
		if !ok {
			logger.Warnf("filter %q: expected a string result, got %T", name, output)

			dynFilters[name] = fmt.Sprint(output)
		}

		// Don't update the cache, since we don't decode the output to the right type.
	}

	return dynFilters, nil
}

func (exprMan *expressionManager) signedURLParams(ctx context.Context, img image, paramsExprName string) (map[string]any, error) {
	paramsExpr, found := exprMan.exprs[img.imgGroup][img.imgType][paramsExprName]
	if !found {
		return nil, fmt.Errorf("%w %q", errMissingExpression, paramsExprName)
	}

	env := types.ExprEnv{
		Ctx:   ctx,
		Files: exprMan.valueMap2FilesMap(img),
		Exprs: exprMan.exprs[img.imgGroup][img.imgType],
	}
	selectorsSum := dynamicFilesChecksum(env.Files)

	if value, ok := exprMan.getCache(img.bucket, img.s3Key, paramsExprName, selectorsSum); ok {
		return value.(map[string]any), nil //nolint: forcetypeassert
	}

	output, err := expr.Run(paramsExpr, env)
	if err != nil {
		return nil, fmt.Errorf("expr: %w", err)
	}

	outputMap, ok := output.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%w: want map[string]any, got %T", errUnexpectedOutputType, output)
	}

	exprMan.updateCache(img.bucket, img.s3Key, paramsExprName, selectorsSum, outputMap)

	return outputMap, nil
}

func (exprMan *expressionManager) valueMap2FilesMap(img image) map[string]types.DynamicInputFile {
	result := make(map[string]types.DynamicInputFile, len(img.dynamicInputFiles)+1)

	result[types.ObjectPreview] = types.DynamicInputFile{
		S3Path:   img.s3Key,
		CacheKey: filepath.Join(exprMan.cacheDir, img.previewCacheKey),
		Date:     img.lastModified,
	}

	for key, value := range img.dynamicInputFiles {
		v := value.value
		v.CacheKey = filepath.Join(exprMan.cacheDir, v.CacheKey)
		result[key] = v
	}

	for _, sel := range exprMan.fileSelectors[img.imgGroup][img.imgType] {
		if _, exists := result[sel]; !exists {
			result[sel] = types.DynamicInputFile{}
		}
	}

	return result
}

func dynamicFilesChecksum(selectors map[string]types.DynamicInputFile) string {
	keys := slices.Sorted(maps.Keys(selectors))

	h := sha256.New()

	for _, k := range keys {
		v := selectors[k]

		// Drop monotonic clock and normalize zone so equal times hash equal.
		t := v.Date.Round(0).UTC()

		h.Write([]byte(k))
		h.Write([]byte{0})
		h.Write([]byte(v.S3Path))
		h.Write([]byte{0})

		var buf [8]byte
		binary.BigEndian.PutUint64(buf[:], uint64(t.UnixNano())) //nolint: gosec // time won't ever be negative
		h.Write(buf[:])

		h.Write([]byte{0xFF}) // entry delimiter
	}

	return hex.EncodeToString(h.Sum(nil))
}
