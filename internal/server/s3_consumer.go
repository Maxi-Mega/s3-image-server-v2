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

type eventConsumer struct {
	productsCfg          config.Products
	cacheRetentionPeriod time.Duration
	cache                *cache
	s3Chan               chan s3.Event
	baseDirChan          chan string
	ootChan              chan s3Event
}

func newS3Consumer(cfg config.Config, cache *cache, s3Chan chan s3.Event) *eventConsumer {
	return &eventConsumer{
		productsCfg: cfg.Products,
		cache:       cache,
		s3Chan:      s3Chan,
		baseDirChan: make(chan string, chanBuf),
		ootChan:     make(chan s3Event, chanBuf),
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

	newObjectTemporizer(consumer.baseDirChan, consumer.ootChan, consumer.cache).goTemporize(ctx)
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
		event.ObjectType, found = consumer.getObjectType(event.ObjectKey, imgType)
		if !found {
			return
		}
	}

	evt := s3Event{
		Event:    event,
		imgGroup: imgGroup,
		imgType:  imgType,
	}

	if event.ObjectType == types.ObjectPreview {
		basePath, ok := utils.GetRegexMatchGroup(consumer.productsCfg.TargetRelativeRgx, event.ObjectKey, 1)
		if !ok {
			logger.Tracef("Preview %s/%q doesn't match the targetRelativeRegexp", event.Bucket, event.ObjectKey)

			return
		}

		evt.baseDir = basePath

		consumer.cache.handleEvent(ctx, evt)
		consumer.baseDirChan <- basePath
	} else {
		consumer.ootChan <- evt
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

func (consumer *eventConsumer) getObjectType(objectKey string, imgType config.ImageType) (objectType types.ObjectType, found bool) {
	cfg := consumer.productsCfg

	switch path.Base(objectKey) {
	case cfg.PreviewFilename:
		return types.ObjectPreview, true
	case cfg.GeonamesFilename:
		return types.ObjectGeonames, true
	case cfg.LocalizationFilename:
		return types.ObjectLocalization, true
	}

	switch {
	case cfg.AdditionalProductFilesRgx.MatchString(objectKey):
		return types.ObjectAdditional, true
	case cfg.FeaturesExtensionRgx.MatchString(objectKey):
		return types.ObjectFeatures, true
	case strings.HasPrefix(objectKey, imgType.ProductPrefix) && cfg.TargetRelativeRgx.MatchString(objectKey):
		return types.ObjectTarget, true
	case strings.HasSuffix(objectKey, cfg.FullProductExtension) && cfg.FullProductSignedURL:
		return types.ObjectFullProduct, true
	}

	return "", false
}
