package types //nolint: revive,nolintlint

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

type CachedObject struct {
	LastModified time.Time `json:"lastModified"`
	CacheKey     string    `json:"cacheKey"`
}

type ProductInformation struct {
	Title    string   `json:"title"`
	Subtitle string   `json:"subtitle"`
	Entries  []string `json:"entries"`
	Summary  string   `json:"summary"`
}

type ImageSize struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

type ImageSummary struct {
	Bucket      string              `json:"bucket"`
	Key         string              `json:"key"`
	Name        string              `json:"name"`
	Group       string              `json:"group"`
	Type        string              `json:"type"`
	Geonames    *Geonames           `json:"geonames"`
	ProductInfo *ProductInformation `json:"productInfo"`
	// Contains the cache key to the image preview.
	CachedObject CachedObject `json:"cachedObject"`
	Size         ImageSize    `json:"size"`
}

type Image struct {
	ImageSummary ImageSummary
	Localization *Localization
	// CachedFileLinks is a map[filename] -> cache key
	CachedFileLinks map[string]string
	// SignedURLs is a map[filename] -> cache key
	SignedURLs map[string]string
	// TargetFiles is a slice of cache keys
	TargetFiles []string
}

type ObjectType string

const (
	ObjectPreview      = "preview"
	ObjectTarget       = "target"
	ObjectDynamicInput = "dynamic_input"
	// ObjectNotYetAssigned represents a special case where
	// we don't have enough information yet to classify the object.
	ObjectNotYetAssigned = "not_yet_assigned"
)

// AllImageSummaries is a map[group] -> map[type] -> images.
type AllImageSummaries map[string]map[string][]ImageSummary

type Cache interface {
	GetAllImages(ctx context.Context, start, end time.Time) AllImageSummaries
	GetImage(ctx context.Context, bucket, name string) (Image, error)
	GetCachedObject(cacheKey string) ([]byte, error)
	DumpImages() map[string][]string
}

type EventType string

const (
	EventReset   = "Reset"
	EventCreated = "ObjectCreated"
	EventRemoved = "ObjectRemoved"
)

type OutEvent struct {
	EventType   EventType  `json:"eventType"`
	ObjectType  ObjectType `json:"objectType"`
	ImageBucket string     `json:"imageBucket"`
	ImageKey    string     `json:"imageKey"`
	ObjectTime  time.Time  `json:"objectTime"`
	// Only filled for EventCreated
	Object any `json:"object,omitempty"`
	// Eventual error
	Error string `json:"error,omitempty"`
}

func (evt OutEvent) String() string {
	if evt.EventType == EventReset {
		return "[Reset]"
	}

	str := fmt.Sprintf("[%s] %s (%s): %q", evt.EventType, evt.ObjectTime, evt.ObjectType, evt.ImageKey)

	if evt.Error != "" {
		str += fmt.Sprintf(" /!\\ error: %v /!\\", evt.Error)
	}

	return str
}

func (evt OutEvent) JSON() ([]byte, error) {
	data, err := json.Marshal(evt)
	if err != nil {
		return fmt.Appendf(nil, `{"error": %q}`, err), fmt.Errorf("failed to marshal OutEvent{%s}: %w", evt.String(), err)
	}

	return data, nil
}
