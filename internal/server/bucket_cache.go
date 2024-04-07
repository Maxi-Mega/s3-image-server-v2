package server

import (
	"context"
	"errors"
	"os"
	"path"
	"path/filepath"
	"sync"
	"time"

	"github.com/Maxi-Mega/s3-image-server-v2/config"
	"github.com/Maxi-Mega/s3-image-server-v2/internal/logger"
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
	case types.ObjectLocalization:
		if img.localization != nil && !img.localization.LastModified.Before(event.ObjectLastModified) {
			return nil
		}
	case types.ObjectAdditional:
		if !img.additional[path.Base(event.ObjectKey)].lastUpdate.Before(event.ObjectLastModified) {
			return nil
		}

		subDir = additionalDirName
	case types.ObjectFeatures:
		if img.features != nil && !img.features.LastModified.Before(event.ObjectLastModified) {
			return nil
		}
	case types.ObjectTarget:
		if !img.targets[path.Base(event.ObjectKey)].lastUpdate.Before(event.ObjectLastModified) {
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
		img.additional = make(map[string]withLastUpdate)
		img.targets = make(map[string]withLastUpdate)
		img.fullProducts = make(map[string]withLastUpdate)

		bc.setDropTimer(event.baseDir, event.Time)
	}

	fullFilePath := filepath.Join(bc.dirPath, img.name, subDir, path.Base(event.ObjectKey))
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

	err := bc.applyObjectTypeSpecificHooks(ctx, event, &img, fullFilePath)
	if err != nil {
		if errors.Is(err, errObjectAlreadyCached) {
			return nil
		}

		logger.Warnf("Something gone wrong with object %q: %v", event.ObjectKey, err)
	}

	bc.images[event.baseDir] = img

	return &types.OutEvent{
		EventType:  types.EventCreated,
		ObjectType: event.ObjectType,
		ImageName:  img.name,
		CacheKey:   bc.getCacheKey(img.name, subDir, event.ObjectKey),
		ObjectTime: event.ObjectLastModified,
	}
}

func (bc *bucketCache) applyObjectTypeSpecificHooks(ctx context.Context, event s3Event, img *image, fullFilePath string) error {
	objKeyBase := path.Base(event.ObjectKey)
	cacheKey := func(subDir string) string { return bc.getCacheKey(img.name, subDir, event.ObjectKey) }

	var err error

	switch event.ObjectType {
	case types.ObjectPreview:
		img.lastModified = event.ObjectLastModified
		img.previewCacheKey = cacheKey("")
	case types.ObjectGeonames:
		img.geonames, err = parseGeonames(fullFilePath, event.ObjectLastModified, cacheKey(""))
	case types.ObjectLocalization:
		img.localization, err = parseLocalization(fullFilePath, event.ObjectLastModified, cacheKey(""))
	case types.ObjectAdditional:
		img.additional[objKeyBase] = withLastUpdate{
			value:      cacheKey(additionalDirName),
			lastUpdate: event.ObjectLastModified,
		}
	case types.ObjectFeatures:
		img.features, err = parseFeatures(bc.cfg.Products, fullFilePath, event.ObjectLastModified, cacheKey(""), event.ObjectKey)
	case types.ObjectTarget:
		img.targets[objKeyBase] = withLastUpdate{
			value:      cacheKey(targetsDirName),
			lastUpdate: event.ObjectLastModified,
		}
	case types.ObjectFullProduct:
		if fullProduct, exists := img.fullProducts[event.ObjectKey]; exists {
			if !fullProduct.lastUpdate.Before(event.ObjectLastModified) {
				return errObjectAlreadyCached
			}
		}

		if signedURL, err := bc.s3Client.GenerateSignedURL(ctx, event.Bucket, event.ObjectKey); err == nil {
			img.fullProducts[event.ObjectKey] = withLastUpdate{
				value:      signedURL,
				lastUpdate: event.ObjectLastModified,
			}
		}
	}

	return err
}

func (bc *bucketCache) handleRemoveEvent(_ context.Context, event s3Event, img image) *types.OutEvent {
	objKeyBase := path.Base(event.ObjectKey)

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
		delete(img.additional, objKeyBase)

		subDir = additionalDirName
	case types.ObjectFeatures:
		img.features = nil
	case types.ObjectTarget:
		delete(img.targets, objKeyBase)

		subDir = targetsDirName
	case types.ObjectFullProduct:
		delete(img.fullProducts, event.ObjectKey)

		deleteFile = false
	}

	if deleteFile {
		fullFilePath := filepath.Join(bc.dirPath, img.name, subDir, objKeyBase)
		if err := os.Remove(fullFilePath); err != nil && !os.IsNotExist(err) {
			logger.Errorf("Failed to delete %q: %v", fullFilePath, err)
		}
	}

	bc.images[event.baseDir] = img

	return &types.OutEvent{
		EventType:  types.EventRemoved,
		ObjectType: event.ObjectType,
		ImageName:  img.name,
		ObjectTime: event.ObjectLastModified,
	}
}

func (bc *bucketCache) getCacheKey(imgName, subDir, objectKey string) string {
	return filepath.Clean(filepath.Join(bc.bucket, imgName, subDir, path.Base(objectKey)))
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
