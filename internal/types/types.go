package types

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/Maxi-Mega/s3-image-server-v2/internal/logger"
)

type CachedObject struct {
	LastModified time.Time `json:"lastModified"`
	CacheKey     string    `json:"cacheKey"`
}

type ImageSummary struct {
	Bucket   string    `json:"bucket"`
	Key      string    `json:"key"`
	Name     string    `json:"name"`
	Group    string    `json:"group"`
	Type     string    `json:"type"`
	Features *Features `json:"features"`
	// Contains the cache key to the image preview.
	CachedObject CachedObject `json:"cachedObject"`
}

type Image struct {
	ImageSummary ImageSummary
	Geonames     *Geonames
	Localization *Localization
	// AdditionalFiles is a map[filename] -> cache key
	AdditionalFiles map[string]string
	// TargetFiles is a slice of cache keys
	TargetFiles []string
	// FullProductFiles is a map[filename] -> cache key
	FullProductFiles map[string]string
}

type ObjectType string

const (
	ObjectPreview      = "preview"
	ObjectGeonames     = "geonames"
	ObjectLocalization = "localization"
	ObjectAdditional   = "additional"
	ObjectFeatures     = "features"
	ObjectTarget       = "target"
	ObjectFullProduct  = "full_product"
	// ObjectNotYetAssigned represents a special case where
	// we don't have enough information yet to classify the object.
	ObjectNotYetAssigned = "not_yet_assigned"
)

// AllImageSummaries is a map[group] -> map[type] -> images.
type AllImageSummaries map[string]map[string][]ImageSummary

type Cache interface {
	GetAllImages(start, end time.Time) AllImageSummaries
	GetImage(bucket, name string) (Image, error)
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

func (evt OutEvent) JSON() []byte {
	data, err := json.Marshal(evt)
	if err != nil {
		logger.Errorf("Failed to marshal OutEvent{%s}: %v", evt.String(), err)

		return []byte(fmt.Sprintf(`{"error": %q`, err))
	}

	return data
}
