package server

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/Maxi-Mega/s3-image-server-v2/config"
	"github.com/Maxi-Mega/s3-image-server-v2/internal/logger"
	"github.com/Maxi-Mega/s3-image-server-v2/internal/observability"
	"github.com/Maxi-Mega/s3-image-server-v2/internal/s3"
	"github.com/Maxi-Mega/s3-image-server-v2/internal/types"
)

var (
	errNoBucketSpecified = errors.New("no bucket name specified")
	errBucketNotFound    = errors.New("can't find bucket")
)

type Server struct {
	cfg      config.Config
	gatherer *observability.Metrics
	buckets  []string

	s3Client s3.Client
	s3Chan   chan s3.Event
	outChan  chan types.OutEvent
	cache    *cache
}

func New(cfg config.Config, gatherer *observability.Metrics) (*Server, error) {
	s3Client, err := s3.NewClient(cfg, gatherer)
	if err != nil {
		return nil, err //nolint:wrapcheck
	}

	buckets := make([]string, 0)

	for _, group := range cfg.Products.ImageGroups {
		bucket := group.Bucket
		if bucket == "" {
			return nil, fmt.Errorf("%w for image group %q", errNoBucketSpecified, group.GroupName)
		}

		if !slices.Contains(buckets, bucket) {
			buckets = append(buckets, bucket)
		}
	}

	outEvents := make(chan types.OutEvent)

	cache, err := newCache(cfg, s3Client, outEvents, gatherer)
	if err != nil {
		return nil, err
	}

	return &Server{
		cfg:      cfg,
		gatherer: gatherer,
		buckets:  buckets,
		s3Client: s3Client,
		s3Chan:   make(chan s3.Event),
		outChan:  outEvents,
		cache:    cache,
	}, nil
}

func (srv *Server) Start(ctx context.Context) (types.Cache, chan types.OutEvent, error) {
	for _, bucket := range srv.buckets {
		ok, err := srv.s3Client.BucketExists(ctx, bucket)
		if err != nil {
			return nil, nil, err //nolint:wrapcheck
		}

		if !ok {
			return nil, nil, fmt.Errorf("%w %q", errBucketNotFound, bucket)
		}
	}

	newS3Consumer(srv.cfg, srv.cache, srv.s3Chan).goConsumeEvents(ctx)

	var err error

	switch srv.cfg.S3.Mode {
	case config.S3ModePolling:
		logger.Trace("Starting server in polling mode ...")

		err = srv.startPollingS3(ctx)
	case config.S3ModeEvent:
		logger.Trace("Starting server in notification mode ...")

		err = srv.subscribeToS3(ctx)
	}

	if err != nil {
		return nil, nil, err
	}

	go srv.runSignedURLRegenerationLoop(ctx)

	return srv.cache, srv.outChan, nil
}

func (srv *Server) startPollingS3(ctx context.Context) error {
	pollingPeriod := srv.cfg.S3.PollingPeriod

	logger.Debugf("Starting to poll buckets with a period of %s", pollingPeriod)

	for _, bucket := range srv.buckets {
		logger.Tracef("Starting to poll bucket %q", bucket)

		err := srv.s3Client.PollOnce(ctx, bucket, srv.s3Chan, pollingPeriod)
		if err != nil {
			return err //nolint:wrapcheck
		}

		time.AfterFunc(5*time.Second, func() { srv.cache.updateMetrics(ctx, bucket) })

		go func(bucket string) {
			for {
				select {
				case <-time.After(pollingPeriod):
					t0 := time.Now()

					err := srv.s3Client.PollOnce(ctx, bucket, srv.s3Chan, pollingPeriod)
					if err != nil {
						logger.Errorf("Failed to poll bucket %q: %v", bucket, err)
					}

					if total := time.Since(t0); total > pollingPeriod {
						logger.Warnf("Polling bucket %q took longer than the polling period", bucket)
					}

					go srv.cache.updateMetrics(ctx, bucket)
				case <-ctx.Done():
					logger.Debugf("Context expired, stopping to poll bucket %q", bucket)

					return
				}
			}
		}(bucket)
	}

	return nil
}

func (srv *Server) subscribeToS3(ctx context.Context) error {
	for _, bucket := range srv.buckets {
		err := srv.s3Client.SubscribeToBucket(ctx, bucket, srv.s3Chan)
		if err != nil {
			return err //nolint:wrapcheck
		}

		err = srv.s3Client.PollOnce(ctx, bucket, srv.s3Chan, time.Minute)
		if err != nil {
			return err //nolint:wrapcheck
		}

		time.AfterFunc(5*time.Second, func() { srv.cache.updateMetrics(ctx, bucket) })
	}

	go func() {
		for ctx.Err() == nil {
			select {
			case <-time.After(time.Minute):
				for _, bucket := range srv.buckets {
					srv.cache.updateMetrics(ctx, bucket)
				}
			case <-ctx.Done():
				logger.Debug("Context expired, closing event channel")

				close(srv.s3Chan)

				return
			}
		}
	}()

	return nil
}

type signedURLRegenerationRequest struct {
	paramsExpr         string
	objectLastModified time.Time
	img                image
}

func (srv *Server) runSignedURLRegenerationLoop(ctx context.Context) {
	ticker := time.NewTicker(s3.SignedURLRegenerationPeriod)
	defer ticker.Stop()

	for ctx.Err() == nil {
		select {
		case now := <-ticker.C:
			deadline := now.Add(24 * time.Hour)
			// map[bucket][image][s3Key] -> signedURLRegenerationRequest
			urlsToRenew := make(map[string]map[string]map[string]signedURLRegenerationRequest)

			for bucketName, cache := range srv.cache.buckets {
				urlsToRenew[bucketName] = cache.findSignedURLsToRenew(deadline, s3.SignedURLLifetime)
			}

			newURLs, err := srv.regenerateSignedURLs(ctx, urlsToRenew)
			if err != nil {
				if len(newURLs) == 0 {
					logger.Errorf("Failed to regenerate signed URLs: %v", err)

					continue
				}

				logger.Errorf("Failed to regenerate some signed URLs: %v", err)
			}

			totalURLsCount := srv.applyRegeneratedSignedURLs(newURLs)

			// Increment only after the apply phase so the metric reflects URLs actually updated in cache.
			srv.gatherer.S3SignedURLRegenCounter.Add(float64(totalURLsCount))

			logger.Infof("Regenerated %d signed URLs in %s", totalURLsCount, time.Since(now))
		case <-ctx.Done():
			return
		}
	}
}

func (srv *Server) applyRegeneratedSignedURLs(newURLs map[string]map[string]map[string]valueWithLastUpdate[signedURL]) int {
	var totalURLsCount int

	for bucketName, images := range newURLs {
		bucket, ok := srv.cache.buckets[bucketName]
		if !ok {
			continue
		}

		bucket.l.Lock()

		for imgName, signedURLs := range images {
			img, ok := bucket.images[imgName]
			if !ok {
				continue
			}

			for s3Key, renewedSignedURL := range signedURLs {
				currentSignedURL, ok := img.signedURLs[s3Key]
				if !ok {
					continue
				}

				if !currentSignedURL.lastUpdate.Equal(renewedSignedURL.lastUpdate) {
					continue
				}

				if currentSignedURL.value.generationDate.After(renewedSignedURL.value.generationDate) {
					continue
				}

				img.signedURLs[s3Key] = renewedSignedURL
				totalURLsCount++
			}

			bucket.images[imgName] = img
		}

		bucket.l.Unlock()
	}

	return totalURLsCount
}

func (srv *Server) regenerateSignedURLs(ctx context.Context, urlsToRenew map[string]map[string]map[string]signedURLRegenerationRequest) (map[string]map[string]map[string]valueWithLastUpdate[signedURL], error) {
	// map[bucket][image][s3Key] -> url
	newURLs := make(map[string]map[string]map[string]valueWithLastUpdate[signedURL])

	var errs []error

	for bucket, images := range urlsToRenew {
		for imgName, s3Keys := range images {
			for s3Key, regenReq := range s3Keys {
				signedURLGenReq := signedURLGenerationRequest{
					bucket:              bucket,
					s3Key:               s3Key,
					objLastModified:     regenReq.objectLastModified,
					img:                 regenReq.img,
					injectParams:        regenReq.paramsExpr != "",
					paramsExpr:          regenReq.paramsExpr,
					exprManager:         srv.cache.exprManager,
					fullProductProtocol: srv.cfg.Products.FullProductProtocol,
					fullProductRootURL:  srv.cfg.Products.FullProductRootURL,
				}

				signURL, err := makeSignedURL(ctx, srv.s3Client, signedURLGenReq)
				if err != nil {
					errs = append(errs, fmt.Errorf("creating signed URL for file %q in bucket %q: %w", s3Key, bucket, err))

					continue
				}

				if _, ok := newURLs[bucket]; !ok {
					newURLs[bucket] = make(map[string]map[string]valueWithLastUpdate[signedURL])
				}

				if _, ok := newURLs[bucket][imgName]; !ok {
					newURLs[bucket][imgName] = make(map[string]valueWithLastUpdate[signedURL])
				}

				newURLs[bucket][imgName][s3Key] = signURL
			}
		}
	}

	return newURLs, errors.Join(errs...)
}
