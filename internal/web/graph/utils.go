package graph

import (
	"errors"
	"fmt"

	"github.com/Maxi-Mega/s3-image-server-v2/config"
	"github.com/Maxi-Mega/s3-image-server-v2/internal/web/graph/model"
)

var (
	errInvalidTimeRange = errors.New("invalid time range")
	errNotFound         = errors.New("not found")
)

func toMapStringAny[V any](m map[string]V) map[string]any {
	result := make(map[string]any, len(m))

	for k, v := range m {
		result[k] = v
	}

	return result
}

func convertDynamicData(dynData config.DynamicData) (*model.DynamicData, error) {
	fileSelectors := make(map[string]model.FileSelector, len(dynData.FileSelectors))
	expressions := make(map[string]string, len(dynData.ExpressionsPrograms))

	for name, sel := range dynData.FileSelectors {
		var kind string

		switch sel.Kind {
		case config.FileSelectorKindCached, config.FileSelectorKindSignedURL:
			kind = sel.Kind
		case config.FileSelectorKindFullProductSignedURL:
			kind = fmt.Sprintf("%s(%s)", config.FileSelectorKindFullProductSignedURL, sel.KindParams[0])
		}

		fileSelectors[name] = model.FileSelector{
			Regex: sel.Rgx.String(),
			Kind:  kind,
			Link:  sel.Link,
		}
	}

	for name, prgm := range dynData.ExpressionsPrograms {
		expressions[name] = prgm.Node().String()
	}

	return &model.DynamicData{
		FileSelectors: fileSelectors,
		Expressions:   expressions,
	}, nil
}
