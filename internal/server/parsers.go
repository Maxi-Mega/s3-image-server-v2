package server

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Maxi-Mega/s3-image-server-v2/config"
	"github.com/Maxi-Mega/s3-image-server-v2/internal/logger"
	"github.com/Maxi-Mega/s3-image-server-v2/internal/types"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func parseFeatures(productsCfg config.Products, filePath string, objDate time.Time, cacheKey string, objKey string) (*types.Features, error) {
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("file %q not found", filePath)
		}

		return nil, err //nolint:wrapcheck
	}

	var rawFeatures types.RawFeaturesFile

	err = json.Unmarshal(fileContent, &rawFeatures)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal from json the content of the features file %q: %w", filePath, err)
	}

	features := types.Features{
		Objects: make(map[string]int),
		CachedObject: types.CachedObject{
			LastModified: objDate,
			CacheKey:     cacheKey,
		},
	}

	for i, rawFeature := range rawFeatures.Features {
		category, ok := parseFeatureStrProp(productsCfg.FeaturesCategoryName, rawFeature.Properties, i, objKey)
		if !ok {
			continue
		}

		class, ok := parseFeatureStrProp(productsCfg.FeaturesClassName, rawFeature.Properties, i, objKey)
		if !ok {
			continue
		}

		features.Class = class
		features.Count++
		features.Objects[category]++
	}

	return &features, nil
}

func parseFeatureStrProp(key string, props map[string]any, idx int, objKey string) (string, bool) {
	propName, ok := props[key]
	if !ok {
		logger.Warnf("Feature n°%d has no %s (object %q)", idx+1, key, objKey)

		return "", false
	}

	rawProp, ok := propName.(string)
	if !ok {
		logger.Warnf("Feature n°%d %s is not a string ('%v') (object %q)", idx+1, key, propName, objKey)

		return "", false
	}

	value := cases.Title(language.English).String(rawProp)
	value = strings.ReplaceAll(value, "_", " ")

	return value, true
}

func parseGeonames(filePath string, objDate time.Time, cacheKey string) (*types.Geonames, error) {
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("file %q not found", filePath)
		}

		return nil, err //nolint:wrapcheck
	}

	var geonames types.Geonames

	err = json.Unmarshal(fileContent, &geonames.Objects)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal from json the content of the geonames file %q: %w", filePath, err)
	}

	geonames.LastModified = objDate
	geonames.CacheKey = cacheKey

	geonames.Sort()

	return &geonames, nil
}

func parseLocalization(filePath string, objDate time.Time, cacheKey string) (*types.Localization, error) {
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("file %q not found", filePath)
		}

		return nil, err //nolint:wrapcheck
	}

	var localization types.Localization

	err = json.Unmarshal(fileContent, &localization)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal from json the content of the localization file %q: %w", filePath, err)
	}

	localization.LastModified = objDate
	localization.CacheKey = cacheKey

	return &localization, nil
}
