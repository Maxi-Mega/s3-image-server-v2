package server

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
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
	l           sync.RWMutex
	s3Client    s3.Client
	exprManager *expressionManager

	bucket  string
	dirPath string
	cfg     config.Config

	images     map[string]image
	dropTimers map[string]*time.Timer
}

func newBucketCache(s3Client s3.Client, exprMan *expressionManager, bucket, dirPath string, cfg config.Config) *bucketCache {
	return &bucketCache{
		s3Client:    s3Client,
		exprManager: exprMan,
		bucket:      bucket,
		dirPath:     dirPath,
		cfg:         cfg,
		images:      make(map[string]image),
		dropTimers:  make(map[string]*time.Timer),
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
		img.s3Key = event.ObjectKey
	case types.ObjectTarget:
		if !img.targets[event.ObjectKey].lastUpdate.Before(event.ObjectLastModified) {
			return nil
		}

		subDir = targetsDirName
	case types.ObjectDynamicInput:
		cachedFile, found := img.dynamicInputFiles[event.InputFile]
		if found {
			if !cachedFile.lastUpdate.Before(event.ObjectLastModified) {
				return nil
			}

			if cachedFile.value.S3Path != event.ObjectKey {
				logger.Errorf("Input file %q for image type %q/%q matches multiple objects; this should be fixed ! (files found so far: %q, %q)", event.InputFile, event.imgGroup.GroupName, event.imgType.Name, cachedFile.value.S3Path, event.ObjectKey)

				return nil
			}
		}

		selector, ok := event.imgType.DynamicData.FileSelectors[event.InputFile]
		if !ok {
			logger.Errorf("Input file %q for image type %q/%q is not defined in the config", event.InputFile, event.imgGroup.GroupName, event.imgType.Name)

			return nil
		}

		switch selector.Kind {
		case config.FileSelectorKindCached:
			subDir = dynamicInputFilesDirName
		case config.FileSelectorKindSignedURL, config.FileSelectorKindFullProductSignedURL:
			download = false
		}
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
		img.baseDir = event.baseDir
		img.bucket = event.Bucket
		img.imgGroup = event.imgGroup.GroupName
		img.imgType = event.imgType.Name
		img.lastModified = event.ObjectLastModified
		img.targets = make(map[string]valueWithLastUpdate[string])
		img.dynamicInputFiles = make(map[string]valueWithLastUpdate[types.DynamicInputFile])
		img.linksFromCache = make(map[string]valueWithLastUpdate[string])
		img.signedURLs = make(map[string]valueWithLastUpdate[signedURL])

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

	eventObj, err := bc.applyObjectTypeSpecificHooks(ctx, event, &img)
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

func (bc *bucketCache) applyObjectTypeSpecificHooks(ctx context.Context, event s3Event, img *image) (eventObj any, err error) {
	cacheKey := func(subDir ...string) string {
		var subdir string

		if len(subDir) > 0 {
			subdir = subDir[0]
		}

		return bc.getCacheKey(img.name, subdir, event.baseDirRelativePath())
	}

	switch event.ObjectType {
	case types.ObjectPreview:
		img.lastModified = event.ObjectLastModified
		img.previewCacheKey = cacheKey()
		eventObj = img.summary(ctx, event.baseDir, bc.cfg.Cache.CacheDir, bc.exprManager)
	case types.ObjectTarget:
		img.targets[event.ObjectKey] = valueWithLastUpdate[string]{
			value:      cacheKey(targetsDirName),
			lastUpdate: event.ObjectLastModified,
		}
		eventObj = img.targets[event.ObjectKey]
	case types.ObjectDynamicInput:
		selector, ok := event.imgType.DynamicData.FileSelectors[event.InputFile]
		if !ok {
			return nil, errNoEventNeeded
		}

		switch selector.Kind {
		case config.FileSelectorKindCached:
			img.dynamicInputFiles[event.InputFile] = valueWithLastUpdate[types.DynamicInputFile]{
				value: types.DynamicInputFile{
					S3Path:   event.ObjectKey,
					CacheKey: cacheKey(dynamicInputFilesDirName),
					Date:     event.ObjectLastModified,
				},
				lastUpdate: event.ObjectLastModified,
			}

			if selector.Link {
				img.linksFromCache[event.ObjectKey] = valueWithLastUpdate[string]{
					value:      cacheKey(dynamicInputFilesDirName),
					lastUpdate: event.ObjectLastModified,
				}
			}
		case config.FileSelectorKindSignedURL:
			if fullProduct, exists := img.signedURLs[event.ObjectKey]; exists {
				// Checking if we have the latest version of the object, and if the signed URL is still valid.
				if !fullProduct.lastUpdate.Before(event.ObjectLastModified) && fullProduct.value.isValid() {
					return nil, errObjectAlreadyCached
				}
			}

			signURL, errURL := bc.s3Client.GenerateSignedURL(ctx, event.Bucket, event.ObjectKey)
			if errURL != nil {
				err = errURL

				break
			}

			img.signedURLs[event.ObjectKey] = valueWithLastUpdate[signedURL]{
				value: signedURL{
					value:          signURL.String(),
					generationDate: time.Now().Truncate(time.Second), // truncating to get closer to the actual generation timestamp
				},
				lastUpdate: event.ObjectLastModified,
			}
		case config.FileSelectorKindFullProductSignedURL:
			if fullProduct, exists := img.signedURLs[event.ObjectKey]; exists {
				// Checking if we have the latest version of the object, and if the signed URL is still valid.
				if !fullProduct.lastUpdate.Before(event.ObjectLastModified) && fullProduct.value.isValid() {
					return nil, errObjectAlreadyCached
				}
			}

			signURL, errURL := bc.s3Client.GenerateSignedURL(ctx, event.Bucket, event.ObjectKey)
			if errURL != nil {
				err = errURL

				break
			}

			img.signedURLs[event.ObjectKey] = valueWithLastUpdate[signedURL]{
				value: signedURL{
					value:          bc.injectFullProductURLParamsFromExpr(ctx, signURL, *img, selector.KindParams[0]),
					generationDate: time.Now().Truncate(time.Second), // truncating to get closer to the actual generation timestamp
				},
				lastUpdate: event.ObjectLastModified,
			}
		}

		eventObj = img.dynamicInputFiles[event.InputFile]
	}

	return eventObj, err
}

func (bc *bucketCache) handleRemoveEvent(_ context.Context, event s3Event, img image) *types.OutEvent {
	var (
		subDir       string
		deleteFile   = true
		updateImages = true
	)

	switch event.ObjectType {
	case types.ObjectPreview:
		if timer, found := bc.dropTimers[img.name]; found {
			if !timer.Stop() {
				return nil // The drop method has already been called.
			}
		}

		bc.dropImage(img.baseDir)

		deleteFile = false // the whole dir has just been deleted
		updateImages = false
	case types.ObjectTarget:
		delete(img.targets, event.ObjectKey)

		subDir = targetsDirName
	case types.ObjectDynamicInput:
		selector, ok := event.imgType.DynamicData.FileSelectors[event.InputFile]
		if !ok {
			return nil
		}

		delete(img.dynamicInputFiles, event.InputFile)

		switch selector.Kind {
		case config.FileSelectorKindCached:
			if selector.Link {
				delete(img.linksFromCache, event.ObjectKey)
			}

			subDir = dynamicInputFilesDirName
		case config.FileSelectorKindSignedURL, config.FileSelectorKindFullProductSignedURL:
			delete(img.signedURLs, event.ObjectKey)

			deleteFile = false
		}
	}

	if deleteFile {
		fullFilePath := filepath.Join(bc.dirPath, img.name, subDir, event.baseDirRelativePath())
		if err := os.Remove(fullFilePath); err != nil && !os.IsNotExist(err) {
			logger.Errorf("Failed to delete %q: %v", fullFilePath, err)
		}
	}

	if updateImages {
		bc.images[event.baseDir] = img
	}

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
	if err := os.RemoveAll(imgDirPath); err != nil && !os.IsNotExist(err) {
		logger.Errorf("Failed to delete %q: %v", imgDirPath, err)
	} else {
		delete(bc.images, imgBaseDir)
		delete(bc.dropTimers, imgBaseDir)
	}
}

func (bc *bucketCache) injectFullProductURLParamsFromExpr(ctx context.Context, signedURL *url.URL, img image, paramsExpr string) string {
	withoutSchemeAndHost := strings.TrimPrefix(signedURL.String(), signedURL.Scheme+"://"+signedURL.Host)
	newURL := bc.cfg.Products.FullProductProtocol + url.QueryEscape(bc.cfg.Products.FullProductRootURL+withoutSchemeAndHost)

	su, err := url.Parse(newURL)
	if err != nil {
		logger.Warnf("Failed to parse full product signed URL %q: %v", signedURL, err)

		return newURL
	}

	params, err := bc.exprManager.signedURLParams(ctx, img, paramsExpr)
	if err != nil {
		logger.Warnf("Failed to get parameters from expression %q for full product signed URL %q: %v", paramsExpr, signedURL, err)

		return newURL
	}

	q := su.Query()

	for name, value := range params {
		q.Add(name, fmt.Sprint(value))
	}

	su.RawQuery = q.Encode()

	return su.String()
}

func (bc *bucketCache) updateMetrics(ctx context.Context, gatherer *observability.Metrics) {
	entries := make(map[string]int, len(bc.images))
	entryLabels := make(map[string][]string, len(bc.images))

	var ok bool

	func() {
		bc.l.RLock()
		defer bc.l.RUnlock()

		for _, img := range bc.images {
			labels, err := bc.exprManager.promLabels(ctx, img)
			if err != nil {
				logger.Warnf("Failed to get Prometheus labels for image %q: %v", img.name, err)

				continue
			}

			labelValues := make([]string, len(bc.cfg.Monitoring.ProductLabels)+1)
			labelValues[0] = bc.bucket

			for i, label := range bc.cfg.Monitoring.ProductLabels {
				labelValues[i+1], ok = labels[label].(string)
				if !ok {
					logger.Warnf("Unexpected type for Prometheus label %q of image %q: want string, got %T", label, img.name, labels[label])
				}
			}

			entries[strings.Join(labelValues, "+")]++
			entryLabels[strings.Join(labelValues, "+")] = labelValues
		}
	}()

	for entry, count := range entries {
		gatherer.CacheImagesPerBucket.WithLabelValues(entryLabels[entry]...).Set(float64(count))
	}

	var (
		cacheFilesCount     int
		cacheFilesSizeBytes int64
	)

	t0 := time.Now()

	err := filepath.WalkDir(bc.dirPath, func(_ string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return fmt.Errorf("can't get info: %w", err)
		}

		cacheFilesCount++
		cacheFilesSizeBytes += info.Size()

		return nil
	})
	if err != nil {
		logger.Warnf("Failed to gather cache files stats in %q: %v", bc.dirPath, err)

		return
	}

	logger.Tracef("Gathered cache stats for bucket %s in %v: %d files, %d bytes", bc.bucket, time.Since(t0), cacheFilesCount, cacheFilesSizeBytes)

	gatherer.CacheFilesPerBucket.WithLabelValues(bc.bucket).Set(float64(cacheFilesCount))
	gatherer.CacheSizePerBucket.WithLabelValues(bc.bucket).Set(float64(cacheFilesSizeBytes))
}
