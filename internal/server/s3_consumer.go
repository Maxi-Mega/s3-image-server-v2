package server

import (
	"context"
	"path"
	"strings"
	"time"

	"github.com/Maxi-Mega/s3-image-server-v2/config"
	"github.com/Maxi-Mega/s3-image-server-v2/internal/logger"
	"github.com/Maxi-Mega/s3-image-server-v2/internal/s3"
	"github.com/Maxi-Mega/s3-image-server-v2/internal/types"
	"github.com/Maxi-Mega/s3-image-server-v2/utils"
)

const chanBuf = 256

type s3Event struct {
	s3.Event
	baseDir  string
	imgGroup config.ImageGroup
	imgType  config.ImageType
}

func (evt s3Event) baseDirRelativePath() string {
	relativePath := strings.TrimPrefix(evt.ObjectKey, evt.baseDir)

	return utils.FormatDirName(relativePath)
}

type eventConsumer struct {
	productsCfg          config.Products
	cacheRetentionPeriod time.Duration
	cache                *cache
	s3Chan               chan s3.Event
	baseDirChan          chan string
	temporizationChan    chan s3Event
}

func newS3Consumer(cfg config.Config, cache *cache, s3Chan chan s3.Event) *eventConsumer {
	return &eventConsumer{
		productsCfg:          cfg.Products,
		cacheRetentionPeriod: cfg.Cache.RetentionPeriod,
		cache:                cache,
		s3Chan:               s3Chan,
		baseDirChan:          make(chan string, chanBuf),
		temporizationChan:    make(chan s3Event, chanBuf),
	}
}

func (consumer *eventConsumer) goConsumeEvents(ctx context.Context) {
	go func() {
		for ctx.Err() == nil {
			select {
			case event, ok := <-consumer.s3Chan:
				if !ok {
					return
				}

				go consumer.processEvent(ctx, event)
			case <-ctx.Done():
				return
			}
		}
	}()

	newObjectTemporizer(consumer.baseDirChan, consumer.temporizationChan, consumer.cache, consumer.productsCfg).goTemporize(ctx)
}

func (consumer *eventConsumer) processEvent(ctx context.Context, event s3.Event) {
	if event.EventType == types.EventCreated {
		if event.ObjectLastModified.Add(consumer.productsCfg.MaxObjectsAge).Before(time.Now()) ||
			time.Until(event.Time.Add(consumer.cacheRetentionPeriod)) < time.Second {
			if event.ObjectType == types.ObjectPreview {
				logger.Debugf("Ignoring image %q because it is %s old", event.ObjectKey, utils.FormatDuration(time.Since(event.ObjectLastModified)))
			}

			return
		}
	}

	imgGroup, imgType, found := consumer.getImageGroupType(event.Bucket, event.ObjectKey)
	if !found {
		return
	}

	if event.ObjectType == "" { // event comes from polling
		event.ObjectType = consumer.getObjectType(event.ObjectKey, imgType)
	}

	evt := s3Event{
		Event:    event,
		imgGroup: imgGroup,
		imgType:  imgType,
	}

	if event.ObjectType == types.ObjectPreview {
		objKeyWithoutPrefix := strings.TrimPrefix(event.ObjectKey, imgType.ProductPrefix)

		basePath, ok := utils.GetRegexNameGroup(imgType.ProductRgx, objKeyWithoutPrefix, "parent")
		if !ok {
			logger.Tracef("Preview %s/%q doesn't match the productRegexp of type %s/%s", event.Bucket, event.ObjectKey, imgGroup.GroupName, imgType.Name)

			return
		}

		basePath = imgType.ProductPrefix + strings.TrimSuffix(basePath, "/")
		evt.baseDir = basePath

		consumer.cache.handleEvent(ctx, evt)
		consumer.baseDirChan <- basePath
	} else {
		consumer.temporizationChan <- evt
	}
}

func (consumer *eventConsumer) getImageGroupType(bucket, objectKey string) (imgGroup config.ImageGroup, imgType config.ImageType, found bool) {
	for _, imgGroup = range consumer.productsCfg.ImageGroups {
		if bucket != imgGroup.Bucket {
			continue
		}

		for _, imgType = range imgGroup.Types {
			if strings.HasPrefix(objectKey, imgType.ProductPrefix) {
				return imgGroup, imgType, true
			}
		}
	}

	return config.ImageGroup{}, config.ImageType{}, false
}

func (consumer *eventConsumer) getObjectType(objectKey string, imgType config.ImageType) types.ObjectType {
	cfg := consumer.productsCfg

	if imgType.PreviewSuffix != "" {
		if strings.HasSuffix(objectKey, imgType.PreviewSuffix) && imgType.ProductRgx.MatchString(strings.TrimPrefix(objectKey, imgType.ProductPrefix)) {
			return types.ObjectPreview
		}
	} else {
		if strings.HasSuffix(objectKey, cfg.DefaultPreviewSuffix) && imgType.ProductRgx.MatchString(strings.TrimPrefix(objectKey, imgType.ProductPrefix)) {
			return types.ObjectPreview
		}
	}

	switch path.Base(objectKey) {
	case cfg.GeonamesFilename:
		return types.ObjectGeonames
	case cfg.LocalizationFilename:
		return types.ObjectLocalization
	}

	switch {
	case cfg.AdditionalProductFilesRgx.MatchString(objectKey):
		return types.ObjectAdditional
	case cfg.FeaturesExtensionRgx.MatchString(objectKey):
		return types.ObjectFeatures
	case strings.HasSuffix(objectKey, cfg.FullProductExtension) && cfg.FullProductSignedURL:
		return types.ObjectFullProduct
	default:
		return types.ObjectNotYetAssigned
	}
}
