package graph

import (
	"encoding/json"
	"errors"
	"io"

	"github.com/Maxi-Mega/s3-image-server-v2/internal/logger"
	"github.com/Maxi-Mega/s3-image-server-v2/internal/types"

	"github.com/99designs/gqlgen/graphql"
)

var errUnsupportedGraphQLOperation = errors.New("unsupported graphQL operation")

func MarshalAllImageSummaries(obj types.AllImageSummaries) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		err := json.NewEncoder(w).Encode(obj)
		if err != nil {
			logger.Error("[graphql] Failed to marshal AllImageSummaries: ", err)
		}
	})
}

func UnmarshalAllImageSummaries(_ any) (types.AllImageSummaries, error) {
	return nil, errUnsupportedGraphQLOperation
}

func MarshalGeonamesObject(obj types.GeonamesObject) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		err := json.NewEncoder(w).Encode(obj)
		if err != nil {
			logger.Error("[graphql] Failed to marshal Geonames: ", err)
		}
	})
}

func UnmarshalGeonamesObject(_ any) (types.GeonamesObject, error) {
	return types.GeonamesObject{}, errUnsupportedGraphQLOperation
}

func MarshalLocalizationCorner(obj types.LocalizationCorner) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		err := json.NewEncoder(w).Encode(obj)
		if err != nil {
			logger.Error("[graphql] Failed to marshal LocalizationCorner: ", err)
		}
	})
}

func UnmarshalLocalizationCorner(_ any) (types.LocalizationCorner, error) {
	return types.LocalizationCorner{}, errUnsupportedGraphQLOperation
}
