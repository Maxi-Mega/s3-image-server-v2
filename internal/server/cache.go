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
	targetsDirName           = "__targets__"
	dynamicInputFilesDirName = "__dynamic_input_files__"
)

var (
	errObjectAlreadyCached = errors.New("already in cache")
	errNoEventNeeded       = errors.New("no event needed")
)

type cache struct {
	gatherer    *observability.Metrics
	cacheDir    string
	buckets     map[string]*bucketCache
	outEvents   chan types.OutEvent
	exprManager *expressionManager
}

type image struct {
	lastModified      time.Time
	bucket, s3Key     string
	name              string
	imgGroup, imgType string
	// map[s3 key] -> cache key & last update
	targets map[string]valueWithLastUpdate[string]
	// map[config key] -> cache key & last update
	dynamicInputFiles map[string]valueWithLastUpdate[types.DynamicInputFile]
	// map[s3 key] -> cache key & last update
	linksFromCache map[string]valueWithLastUpdate[string]
	// map[s3 key] -> signed URL & last update
	signedURLs      map[string]valueWithLastUpdate[signedURL]
	previewCacheKey string
}

// summary returns the [ImageSummary] of this [image],
// the name parameter corresponds to the image base dir.
func (img image) summary(ctx context.Context, name, cacheDir string, exprMan *expressionManager) types.ImageSummary {
	var displayName string

	geonames, err := exprMan.imageGeonames(ctx, img)
	if err != nil {
		logger.Errorf("Failed to evaluate image geonames for %q: %v", img.name, err)

		displayName = "No geonames found"
	} else if geonames != nil {
		displayName = geonames.GetTopLevel()
	}

	productInfo, err := exprMan.productInfo(ctx, img)
	if err != nil {
		logger.Errorf("Failed to evaluate product information for %q: %v", img.name, err)
	}

	imgSize, err := utils.GetImageSize(img.previewCacheKey, cacheDir)
	if err != nil {
		logger.Warnf("Failed to get image size of %q: %v", img.previewCacheKey, err)
	}

	return types.ImageSummary{
		Bucket:      img.bucket,
		Key:         name,
		Name:        displayName,
		Group:       img.imgGroup,
		Type:        img.imgType,
		Geonames:    geonames,
		ProductInfo: productInfo,
		CachedObject: types.CachedObject{
			LastModified: img.lastModified,
			CacheKey:     img.previewCacheKey,
		},
		Size: imgSize,
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

	exprManager := newExpressionManager(cfg)

	buckets := make(map[string]*bucketCache)

	for _, group := range cfg.Products.ImageGroups {
		bucket := group.Bucket
		if _, found := buckets[bucket]; !found {
			dir := filepath.Join(cfg.Cache.CacheDir, bucket)

			logger.Tracef("Creating cache dir for bucket %q at %s", bucket, dir)

			if err = os.Mkdir(dir, 0700); err != nil {
				return nil, fmt.Errorf("can't create cache dir: %w", err)
			}

			buckets[group.Bucket] = newBucketCache(s3Client, exprManager, bucket, dir, cfg)
		}
	}

	return &cache{
		gatherer:    gatherer,
		cacheDir:    cfg.Cache.CacheDir,
		buckets:     buckets,
		outEvents:   outChan,
		exprManager: exprManager,
	}, nil
}

func (c *cache) GetAllImages(ctx context.Context, start, end time.Time) types.AllImageSummaries {
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

			allImages[grp][typ] = append(allImages[grp][typ], img.summary(ctx, name, c.cacheDir, c.exprManager))
		}

		bucket.l.RUnlock()
	}

	return allImages
}

func (c *cache) GetImage(ctx context.Context, bucketName, name string) (types.Image, error) {
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

	localization, err := c.exprManager.imageLocalization(ctx, img)
	if err != nil {
		logger.Errorf("Failed to evaluate image localization for %q: %v", img.name, err)
	}

	return types.Image{
		ImageSummary:    img.summary(ctx, name, c.cacheDir, c.exprManager),
		Localization:    localization,
		CachedFileLinks: toFilenameValueMap(img.linksFromCache),
		SignedURLs:      toFilenameValueMap(img.signedURLs),
		TargetFiles:     targetFiles,
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
