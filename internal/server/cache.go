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
	"github.com/Maxi-Mega/s3-image-server-v2/internal/s3"
	"github.com/Maxi-Mega/s3-image-server-v2/internal/types"
	"github.com/Maxi-Mega/s3-image-server-v2/utils"
)

const (
	additionalDirName = "__additional__"
	targetsDirName    = "__targets__"
)

var errObjectAlreadyCached = errors.New("already in cache")

type cache struct {
	cacheDir  string
	buckets   map[string]*bucketCache
	outEvents chan types.OutEvent
}

type image struct {
	lastModified        time.Time
	bucket, s3Key       string
	name                string
	imgGroup, imgType   string
	geonames            *types.Geonames
	localization        *types.Localization
	features            *types.Features
	additional, targets map[string]time.Time
	fullProducts        map[string]fullProduct
	previewCacheKey     string
}

func (img image) summary() types.ImageSummary {
	return types.ImageSummary{
		Bucket: img.bucket,
		Name:   img.name,
		Group:  img.imgGroup,
		Type:   img.imgType,
		CachedObject: types.CachedObject{
			LastModified: img.lastModified,
			CacheKey:     img.previewCacheKey,
		},
	}
}

type fullProduct struct {
	signedURL  string
	lastUpdate time.Time
}

func newCache(cfg config.Config, s3Client s3.Client, outChan chan types.OutEvent) (*cache, error) {
	logger.Debug("Creating cache dir at ", cfg.Cache.CacheDir)

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
		cacheDir:  cfg.Cache.CacheDir,
		buckets:   buckets,
		outEvents: outChan,
	}, nil
}

func (c *cache) handleEvent(ctx context.Context, event s3Event) {
	bucket, ok := c.buckets[event.Bucket]
	if !ok {
		return
	}

	bucket.l.Lock()
	defer bucket.l.Unlock()

	img, ok := bucket.images[event.baseDir]
	if !ok && event.EventType == types.EventRemoved {
		return
	}

	var outEvent *types.OutEvent

	switch {
	case event.EventType == types.EventCreated:
		outEvent = bucket.handleCreateEvent(ctx, event, img)
	case event.EventType == types.EventRemoved:
		outEvent = bucket.handleRemoveEvent(ctx, event, img)
	default:
		logger.Warnf("Unknown s3 event %q was handed to cache", event.EventType)
	}

	if outEvent != nil {
		c.outEvents <- *outEvent
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
		if strings.HasPrefix(entry, imgBaseDir) {
			return true, imgBaseDir
		}
	}

	return false, ""
}

func (c *cache) GetAllImages() types.AllImageSummaries {
	allImages := make(types.AllImageSummaries)

	for _, bucket := range c.buckets {
		bucket.l.RLock()

		for _, img := range bucket.images {
			grp, typ := img.imgGroup, img.imgType
			if _, ok := allImages[grp]; !ok {
				allImages[grp] = make(map[string][]types.ImageSummary)
			}

			allImages[grp][typ] = append(allImages[grp][typ], img.summary())
		}

		bucket.l.RUnlock()
	}

	return allImages
}

func (c *cache) GetImageDetails(cacheKey string) (types.Image, bool) {
	// TODO implement me
	panic("implement me")
}

func (c *cache) GetCachedObject(cacheKey string) ([]byte, error) {
	return os.ReadFile(filepath.Join(c.cacheDir, cacheKey)) //nolint:wrapcheck
}
