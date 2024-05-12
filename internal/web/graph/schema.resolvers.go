package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.
// Code generated by github.com/99designs/gqlgen version v0.17.45

import (
	"context"
	"errors"
	"time"

	"github.com/Maxi-Mega/s3-image-server-v2/internal/types"
)

// Objects is the resolver for the Objects field.
func (r *featuresResolver) Objects(ctx context.Context, obj *types.Features) (map[string]interface{}, error) {
	return toMapStringAny(obj.Objects), nil
}

// AdditionalFiles is the resolver for the AdditionalFiles field.
func (r *imageResolver) AdditionalFiles(ctx context.Context, obj *types.Image) (map[string]interface{}, error) {
	return toMapStringAny(obj.AdditionalFiles), nil
}

// FullProductFiles is the resolver for the FullProductFiles field.
func (r *imageResolver) FullProductFiles(ctx context.Context, obj *types.Image) (map[string]interface{}, error) {
	return toMapStringAny(obj.FullProductFiles), nil
}

// GetAllImageSummaries is the resolver for the getAllImageSummaries field.
func (r *queryResolver) GetAllImageSummaries(ctx context.Context, from *time.Time, to *time.Time) (types.AllImageSummaries, error) {
	start := time.Date(0, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Now().Add(24 * time.Hour)

	if from != nil {
		start = *from
	}

	if to != nil {
		end = *to
	}

	if start.After(end) {
		return nil, errInvalidTimeRange
	}

	return r.Cache.GetAllImages(start, end), nil
}

// GetImage is the resolver for the getImage field.
func (r *queryResolver) GetImage(ctx context.Context, bucket string, name string) (*types.Image, error) {
	img, err := r.Cache.GetImage(bucket, name)
	if err != nil {
		if errors.Is(err, types.ErrImageNotFound) {
			return nil, nil
		}

		return nil, err
	}

	return &img, nil
}

// Features returns FeaturesResolver implementation.
func (r *Resolver) Features() FeaturesResolver { return &featuresResolver{r} }

// Image returns ImageResolver implementation.
func (r *Resolver) Image() ImageResolver { return &imageResolver{r} }

// Query returns QueryResolver implementation.
func (r *Resolver) Query() QueryResolver { return &queryResolver{r} }

type featuresResolver struct{ *Resolver }
type imageResolver struct{ *Resolver }
type queryResolver struct{ *Resolver }
