package server

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Maxi-Mega/s3-image-server-v2/config"
	"github.com/Maxi-Mega/s3-image-server-v2/internal/logger"
	"github.com/Maxi-Mega/s3-image-server-v2/internal/observability"
	"github.com/Maxi-Mega/s3-image-server-v2/internal/s3"
	"github.com/Maxi-Mega/s3-image-server-v2/internal/types"
	"github.com/Maxi-Mega/s3-image-server-v2/utils"
)

const (
	additionalDirName = "__additional__"
	targetsDirName    = "__targets__"
)

var (
	errObjectAlreadyCached = errors.New("already in cache")
	errNoEventNeeded       = errors.New("no event needed")
)

type cache struct {
	gatherer  *observability.Metrics
	cacheDir  string
	buckets   map[string]*bucketCache
	outEvents chan types.OutEvent
}

type image struct {
	lastModified      time.Time
	bucket, s3Key     string
	name              string
	imgGroup, imgType string
	geonames          *types.Geonames
	localization      *types.Localization
	features          *types.Features
	// map[filename] -> cache key & last update
	additional map[string]valueWithLastUpdate[string]
	// map[s3 key] -> cache key & last update
	targets map[string]valueWithLastUpdate[string]
	// map[filename] -> signed URL & last update
	fullProducts    map[string]valueWithLastUpdate[signedURL]
	previewCacheKey string
}

// summary returns the [ImageSummary] of this [image],
// the name parameter corresponds to the image base dir.
func (img image) summary(name string) types.ImageSummary {
	var displayName string

	if img.geonames != nil {
		displayName = img.geonames.GetTopLevel()
	} else {
		displayName = "No geonames found"
	}

	return types.ImageSummary{
		Bucket:   img.bucket,
		Key:      name,
		Name:     displayName,
		Group:    img.imgGroup,
		Type:     img.imgType,
		Features: img.features,
		CachedObject: types.CachedObject{
			LastModified: img.lastModified,
			CacheKey:     img.previewCacheKey,
		},
	}
}

type valueWithLastUpdate[T any] struct {
	value      T
	lastUpdate time.Time
}

func (v valueWithLastUpdate[T]) String() string {
	return fmt.Sprint(v.value)
}

type signedURL struct {
	value          string
	generationDate time.Time
}

func (su signedURL) String() string {
	return su.value
}

func (su signedURL) isValid() bool {
	return time.Since(su.generationDate) < s3.SignedURLLifetime
}

func newCache(cfg config.Config, s3Client s3.Client, outChan chan types.OutEvent, gatherer *observability.Metrics) (*cache, error) {
	logger.Debug("Using cache dir at ", cfg.Cache.CacheDir)

	err := utils.CreateDir(cfg.Cache.CacheDir)
	if err != nil {
		return nil, fmt.Errorf("can't create cache dir: %w", err)
	}

	buckets := make(map[string]*bucketCache)

	for _, group := range cfg.Products.ImageGroups {
		bucket := group.Bucket
		if _, found := buckets[bucket]; !found {
			dir := filepath.Join(cfg.Cache.CacheDir, bucket)

			logger.Tracef("Creating cache dir for bucket %q at %s", bucket, dir)

			if err = os.Mkdir(dir, 0700); err != nil {
				return nil, fmt.Errorf("can't create cache dir: %w", err)
			}

			buckets[group.Bucket] = newBucketCache(s3Client, bucket, dir, cfg)
		}
	}

	return &cache{
		gatherer:  gatherer,
		cacheDir:  cfg.Cache.CacheDir,
		buckets:   buckets,
		outEvents: outChan,
	}, nil
}

func (c *cache) GetAllImages(start, end time.Time) types.AllImageSummaries {
	allImages := make(types.AllImageSummaries)

	for _, bucket := range c.buckets {
		bucket.l.RLock()

		for name, img := range bucket.images {
			if img.lastModified.Before(start) || img.lastModified.After(end) {
				continue
			}

			grp, typ := img.imgGroup, img.imgType
			if _, ok := allImages[grp]; !ok {
				allImages[grp] = make(map[string][]types.ImageSummary)
			}

			allImages[grp][typ] = append(allImages[grp][typ], img.summary(name))
		}

		bucket.l.RUnlock()
	}

	return allImages
}

func (c *cache) GetImage(bucketName, name string) (types.Image, error) {
	bucket, ok := c.buckets[bucketName]
	if !ok {
		return types.Image{}, types.ErrImageNotFound
	}

	bucket.l.RLock()
	defer bucket.l.RUnlock()

	img, ok := bucket.images[name]
	if !ok {
		return types.Image{}, types.ErrImageNotFound
	}

	targetFiles := make([]string, 0, len(img.targets))

	for _, target := range img.targets {
		targetFiles = append(targetFiles, target.value)
	}

	return types.Image{
		ImageSummary:     img.summary(name),
		Geonames:         img.geonames,
		Localization:     img.localization,
		AdditionalFiles:  toFilenameValueMap(img.additional),
		TargetFiles:      targetFiles,
		FullProductFiles: toFilenameValueMap(img.fullProducts),
	}, nil
}

func (c *cache) GetCachedObject(cacheKey string) ([]byte, error) {
	return os.ReadFile(filepath.Join(c.cacheDir, cacheKey)) //nolint:wrapcheck
}

func (c *cache) DumpImages() map[string][]string {
	imagesPerBucket := make(map[string][]string, len(c.buckets))

	for _, bucket := range c.buckets {
		bucket.l.RLock()

		imagesPerBucket[bucket.bucket] = make([]string, 0, len(bucket.images))

		for imgBaseDir := range bucket.images {
			imagesPerBucket[bucket.bucket] = append(imagesPerBucket[bucket.bucket], imgBaseDir)
		}

		bucket.l.RUnlock()
	}

	return imagesPerBucket
}

func (c *cache) handleEvent(ctx context.Context, event s3Event) {
	bucket, ok := c.buckets[event.Bucket]
	if !ok {
		return
	}

	c.gatherer.S3EventsCounter.WithLabelValues(event.Bucket).Inc()

	bucket.l.Lock()
	defer bucket.l.Unlock()

	img, ok := bucket.images[event.baseDir]
	if !ok && event.EventType == types.EventRemoved {
		return
	}

	var outEvent *types.OutEvent

	switch event.EventType {
	case types.EventCreated:
		outEvent = bucket.handleCreateEvent(ctx, event, img)
	case types.EventRemoved:
		outEvent = bucket.handleRemoveEvent(ctx, event, img)
	default:
		logger.Warnf("Unknown s3 event %q was handed to cache", event.EventType)
	}

	if outEvent != nil {
		c.outEvents <- *outEvent

		bucket.updateMetrics(c.gatherer)
	}
}

func (c *cache) matchesEntry(bucketName string, entry string) (match bool, baseDir string) {
	bucket, ok := c.buckets[bucketName]
	if !ok {
		return false, ""
	}

	bucket.l.RLock()
	defer bucket.l.RUnlock()

	for imgBaseDir := range bucket.images {
		if strings.HasPrefix(entry, imgBaseDir+"/") {
			return true, imgBaseDir
		}
	}

	return false, ""
}

func toFilenameValueMap[T any](m map[string]valueWithLastUpdate[T]) map[string]string {
	result := make(map[string]string, len(m))

	for filename, metadata := range m {
		result[filename] = metadata.String()
	}

	return result
}
