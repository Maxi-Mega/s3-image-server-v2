package server

import (
	"context"
	"errors"
	"net/url"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/Maxi-Mega/s3-image-server-v2/config"
	"github.com/Maxi-Mega/s3-image-server-v2/internal/logger"
	"github.com/Maxi-Mega/s3-image-server-v2/internal/observability"
	"github.com/Maxi-Mega/s3-image-server-v2/internal/s3"
	"github.com/Maxi-Mega/s3-image-server-v2/internal/types"
	"github.com/Maxi-Mega/s3-image-server-v2/utils"
)

type bucketCache struct {
	l        sync.RWMutex
	s3Client s3.Client

	bucket  string
	dirPath string
	cfg     config.Config

	images     map[string]image
	dropTimers map[string]*time.Timer
}

func newBucketCache(s3Client s3.Client, bucket, dirPath string, cfg config.Config) *bucketCache {
	return &bucketCache{
		s3Client:   s3Client,
		bucket:     bucket,
		dirPath:    dirPath,
		cfg:        cfg,
		images:     make(map[string]image),
		dropTimers: make(map[string]*time.Timer),
	}
}

func (bc *bucketCache) handleCreateEvent(ctx context.Context, event s3Event, img image) *types.OutEvent {
	var (
		subDir   string
		download = true
	)

	switch event.ObjectType {
	case types.ObjectPreview:
		if !img.lastModified.Before(event.ObjectLastModified) {
			return nil
		}

		img.lastModified = event.ObjectLastModified
		img.s3Key = event.Bucket
	case types.ObjectGeonames:
		if img.geonames != nil && !img.geonames.LastModified.Before(event.ObjectLastModified) {
			return nil
		}

		// Geonames file may sometime be empty, so we prevent further processing to avoid getting an error like "end of JSON input".
		if event.Size == 0 {
			logger.Warnf("Geonames file %q is empty; ignoring it", event.ObjectKey)

			return nil
		}
	case types.ObjectLocalization:
		if img.localization != nil && !img.localization.LastModified.Before(event.ObjectLastModified) {
			return nil
		}
	case types.ObjectAdditional:
		if !img.additional[event.ObjectKey].lastUpdate.Before(event.ObjectLastModified) {
			return nil
		}

		subDir = additionalDirName
	case types.ObjectFeatures:
		if img.features != nil && !img.features.LastModified.Before(event.ObjectLastModified) {
			return nil
		}
	case types.ObjectTarget:
		if !img.targets[event.ObjectKey].lastUpdate.Before(event.ObjectLastModified) {
			return nil
		}

		subDir = targetsDirName
	case types.ObjectFullProduct:
		if !bc.cfg.Products.FullProductSignedURL {
			return nil
		}

		download = false
	}

	if img.name == "" {
		imgName := utils.FormatDirName(event.baseDir)

		logger.Debugf("Creating dir for image %s/%q at %q", event.Bucket, imgName, filepath.Join(bc.dirPath, imgName))

		err := utils.CreateDir(filepath.Join(bc.dirPath, imgName))
		if err != nil {
			logger.Errorf("Failed to create img cache dir for %q: %v", event.baseDir, err)

			return nil
		}

		img.name = imgName
		img.bucket = event.Bucket
		img.imgGroup = event.imgGroup.GroupName
		img.imgType = event.imgType.Name
		img.lastModified = event.ObjectLastModified
		img.additional = make(map[string]valueWithLastUpdate[string])
		img.targets = make(map[string]valueWithLastUpdate[string])
		img.fullProducts = make(map[string]valueWithLastUpdate[signedURL])

		bc.setDropTimer(event.baseDir, event.Time)
	}

	fullFilePath := filepath.Join(bc.dirPath, img.name, subDir, event.baseDirRelativePath())
	if stat, exists, err := utils.FileStat(fullFilePath); err != nil {
		logger.Errorf("Failed to check for file existence: %v", err)

		return nil
	} else if exists {
		if event.Size == stat.Size() && event.ObjectLastModified.Equal(stat.ModTime()) {
			return nil
		}
	}

	if download {
		err := bc.s3Client.DownloadObject(ctx, event.Bucket, event.ObjectKey, fullFilePath)
		if err != nil {
			logger.Errorf("Failed to download object %q to cache: %v", event.ObjectKey, err)

			return nil
		}
	}

	eventObj, err := bc.applyObjectTypeSpecificHooks(ctx, event, &img, fullFilePath)
	if err != nil {
		switch {
		case errors.Is(err, errObjectAlreadyCached):
			// pass
		case errors.Is(err, errNoEventNeeded):
			// pass
		default:
			logger.Warnf("Something went wrong with object %s/%q: %v", event.Bucket, event.ObjectKey, err)
		}

		return nil
	}

	bc.images[event.baseDir] = img

	return &types.OutEvent{
		EventType:   types.EventCreated,
		ObjectType:  event.ObjectType,
		ImageBucket: img.bucket,
		ImageKey:    event.baseDir,
		ObjectTime:  event.ObjectLastModified,
		Object:      eventObj,
	}
}

func (bc *bucketCache) applyObjectTypeSpecificHooks(ctx context.Context, event s3Event, img *image, fullFilePath string) (eventObj any, err error) {
	cacheKey := func(subDir ...string) string {
		subdir := ""
		if len(subDir) > 0 {
			subdir = subDir[0]
		}

		return bc.getCacheKey(img.name, subdir, event.baseDirRelativePath())
	}

	switch event.ObjectType {
	case types.ObjectPreview:
		img.lastModified = event.ObjectLastModified
		img.previewCacheKey = cacheKey()
		eventObj = img.summary(event.baseDir)
	case types.ObjectGeonames:
		img.geonames, err = parseGeonames(fullFilePath, event.ObjectLastModified, cacheKey())
		if img.geonames != nil {
			eventObj = struct {
				TopLevel string `json:"topLevel"`
			}{
				TopLevel: img.geonames.GetTopLevel(),
			}
		}
	case types.ObjectLocalization:
		img.localization, err = parseLocalization(fullFilePath, event.ObjectLastModified, cacheKey())
		eventObj = img.localization
	case types.ObjectAdditional:
		img.additional[event.ObjectKey] = valueWithLastUpdate[string]{
			value:      cacheKey(additionalDirName),
			lastUpdate: event.ObjectLastModified,
		}
		eventObj = img.additional[event.ObjectKey]
	case types.ObjectFeatures:
		img.features, err = parseFeatures(bc.cfg.Products, fullFilePath, event.ObjectLastModified, cacheKey(), event.ObjectKey)
		eventObj = img.features
	case types.ObjectTarget:
		img.targets[event.ObjectKey] = valueWithLastUpdate[string]{
			value:      cacheKey(targetsDirName),
			lastUpdate: event.ObjectLastModified,
		}
		eventObj = img.targets[event.ObjectKey]
	case types.ObjectFullProduct:
		if fullProduct, exists := img.fullProducts[event.ObjectKey]; exists {
			// Checking if we have the latest version of the object, and if the signed URL is still valid.
			if !fullProduct.lastUpdate.Before(event.ObjectLastModified) && fullProduct.value.isValid() {
				return nil, errObjectAlreadyCached
			}
		}

		if signURL, err := bc.s3Client.GenerateSignedURL(ctx, event.Bucket, event.ObjectKey); err == nil {
			img.fullProducts[event.ObjectKey] = valueWithLastUpdate[signedURL]{
				value: signedURL{
					value:          injectFullProductURLParams(event.imgGroup, event.ObjectKey, signURL),
					generationDate: time.Now().Truncate(time.Second), // truncating to get closer to the actual gen TS
				},
				lastUpdate: event.ObjectLastModified,
			}
			eventObj = img.fullProducts[event.ObjectKey]
		}
	}

	return eventObj, err
}

func (bc *bucketCache) handleRemoveEvent(_ context.Context, event s3Event, img image) *types.OutEvent {
	var (
		subDir     string
		deleteFile = true
	)

	switch event.ObjectType {
	case types.ObjectPreview:
		if timer, found := bc.dropTimers[img.name]; found {
			if !timer.Stop() {
				return nil // The drop method has already been called.
			}
		}

		bc.dropImage(img.name)

		return nil
	case types.ObjectGeonames:
		img.geonames = nil
	case types.ObjectLocalization:
		img.localization = nil
	case types.ObjectAdditional:
		delete(img.additional, event.ObjectKey)

		subDir = additionalDirName
	case types.ObjectFeatures:
		img.features = nil
	case types.ObjectTarget:
		delete(img.targets, event.ObjectKey)

		subDir = targetsDirName
	case types.ObjectFullProduct:
		delete(img.fullProducts, event.ObjectKey)

		deleteFile = false // not a file
	}

	if deleteFile {
		fullFilePath := filepath.Join(bc.dirPath, img.name, subDir, event.baseDirRelativePath())
		if err := os.Remove(fullFilePath); err != nil && !os.IsNotExist(err) {
			logger.Errorf("Failed to delete %q: %v", fullFilePath, err)
		}
	}

	bc.images[event.baseDir] = img

	return &types.OutEvent{
		EventType:   types.EventRemoved,
		ObjectType:  event.ObjectType,
		ImageBucket: img.bucket,
		ImageKey:    event.baseDir,
		ObjectTime:  event.ObjectLastModified,
	}
}

func (bc *bucketCache) getCacheKey(imgName, subDir, filename string) string {
	return filepath.Clean(filepath.Join(bc.bucket, imgName, subDir, filename))
}

func (bc *bucketCache) setDropTimer(baseDir string, cacheAddTime time.Time) {
	if timer, exists := bc.dropTimers[baseDir]; exists {
		timer.Stop()
	}

	expiresIn := time.Until(cacheAddTime.Add(bc.cfg.Products.MaxObjectsAge))
	bc.dropTimers[baseDir] = time.AfterFunc(expiresIn, func() {
		bc.l.Lock()
		defer bc.l.Unlock()

		bc.dropImage(baseDir)
	})
}

func (bc *bucketCache) dropImage(imgBaseDir string) {
	imgDirPath := filepath.Join(bc.dirPath, imgBaseDir)
	if err := os.Remove(imgDirPath); err != nil && !os.IsNotExist(err) {
		logger.Errorf("Failed to delete %q: %v", imgDirPath, err)
	} else {
		delete(bc.images, imgBaseDir)
		delete(bc.dropTimers, imgBaseDir)
	}
}

func (bc *bucketCache) updateMetrics(gatherer *observability.Metrics) {
	gatherer.CacheImagesPerBucket.WithLabelValues(bc.bucket).Set(float64(len(bc.images)))
}

func injectFullProductURLParams(imgGroup config.ImageGroup, fpoKey string, signedURL string) string {
	if imgGroup.FullProductURLParamsRgx == nil {
		return signedURL
	}

	namedGroups, ok := utils.GetAllRegexNameGroups(imgGroup.FullProductURLParamsRgx, fpoKey)
	if !ok {
		logger.Warnf("The fullProductURLParamsRegexp of group %q doesn't match full product file %q", imgGroup.GroupName, fpoKey)

		return signedURL
	}

	su, err := url.Parse(signedURL)
	if err != nil {
		logger.Warnf("Failed to parse signed URL %q: %v", signedURL, err)

		return signedURL
	}

	q := su.Query()

	for _, param := range imgGroup.FullPoductURLParams {
		switch param.Type {
		case config.FullProductURLParamConstant:
			q.Add(param.Name, param.Value)
		case config.FullProductURLParamRegexp:
			value, ok := namedGroups[param.Name]
			if !ok { // should not happen since it's checked at config loading
				logger.Warnf("Param %q not found in fullProductURLParamsRegexp of group %q", param.Name, imgGroup.GroupName)

				continue
			}

			if param.ValueMapping != nil {
				value, ok = param.ValueMapping[value]
				if !ok {
					logger.Warnf("Value mapping for %q was not found in for param %q of group %q", value, param.Name, imgGroup.GroupName)

					value = "UNKNOWN"
				}
			}

			q.Add(param.Name, value)
		}
	}

	su.RawQuery = q.Encode()

	return su.String()
}
