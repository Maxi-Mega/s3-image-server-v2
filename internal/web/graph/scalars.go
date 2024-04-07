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

func MarshalAllImageSummaries(m types.AllImageSummaries) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		err := json.NewEncoder(w).Encode(m)
		if err != nil {
			logger.Error("[graphql] Failed to marshal AllImageSummaries: ", err)
		}
	})
}

func UnmarshalAllImageSummaries(_ any) (types.AllImageSummaries, error) {
	return nil, errUnsupportedGraphQLOperation
}

func MarshalFeatures(f types.Features) graphql.Marshaler {
	return graphql.WriterFunc(func(w io.Writer) {
		err := json.NewEncoder(w).Encode(f)
		if err != nil {
			logger.Error("[graphql] Failed to marshal AllImageSummaries: ", err)
		}
	})
}

func UnmarshalFeatures(_ any) (types.Features, error) {
	return types.Features{}, errUnsupportedGraphQLOperation
}
