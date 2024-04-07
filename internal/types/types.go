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
	Bucket       string `json:"bucket"`
	Name         string `json:"name"`
	Group        string `json:"group"`
	Type         string `json:"type"`
	CachedObject        // inlined
}

type Image struct {
	ImageSummary     // inlined
	Geonames         *Geonames
	Localization     *Localization
	Features         *Features
	AdditionalFiles  map[string]string
	TargetFiles      map[string]string
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
)

// AllImageSummaries is a map[group] -> map[type] -> images
type AllImageSummaries map[string]map[string][]ImageSummary

type Cache interface {
	GetAllImages() AllImageSummaries
	GetImageDetails(cacheKey string) (Image, bool)
	GetCachedObject(cacheKey string) ([]byte, error)
}

type EventType string

const (
	EventCreated = "ObjectCreated"
	EventRemoved = "ObjectRemoved"
)

type OutEvent struct {
	EventType  EventType  `json:"eventType"`
	ObjectType ObjectType `json:"objectType"`
	ImageName  string     `json:"imageName"`
	CacheKey   string     `json:"cacheKey"`
	ObjectTime time.Time  `json:"objectTime"`
	// Eventual error
	Error string `json:"error,omitempty"`
}

func (evt OutEvent) String() string {
	str := fmt.Sprintf("[%s] %s (%s): %q (%s)", evt.EventType, evt.ObjectTime, evt.ObjectType, evt.ImageName, evt.CacheKey)

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
