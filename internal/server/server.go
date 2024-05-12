package server

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/Maxi-Mega/s3-image-server-v2/config"
	"github.com/Maxi-Mega/s3-image-server-v2/internal/logger"
	"github.com/Maxi-Mega/s3-image-server-v2/internal/metrics"
	"github.com/Maxi-Mega/s3-image-server-v2/internal/s3"
	"github.com/Maxi-Mega/s3-image-server-v2/internal/types"
)

var (
	errNoBucketSpecified = errors.New("no bucket name specified")
	errBucketNotFound    = errors.New("can't find bucket")
)

type Server struct {
	cfg      config.Config
	gatherer *metrics.Metrics
	buckets  []string

	s3Client s3.Client
	s3Chan   chan s3.Event
	outChan  chan types.OutEvent
	cache    *cache
}

func New(cfg config.Config, gatherer *metrics.Metrics) (*Server, error) {
	s3Client, err := s3.NewClient(cfg)
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

	if srv.cfg.S3.PollingMode {
		logger.Trace("Starting server in polling mode ...")

		err = srv.startPollingS3(ctx)
	} else {
		logger.Trace("Starting server in notification mode ...")

		err = srv.subscribeToS3(ctx)
	}

	go func() {
		for evt := range srv.outChan {
			logger.Info("Event: ", evt.String()) // TODO: WS
		}
	}()

	return srv.cache, srv.outChan, err
}

func (srv *Server) startPollingS3(ctx context.Context) error {
	pollingPeriod := srv.cfg.S3.PollingPeriod

	logger.Debugf("Starting to poll buckets with a period of %s", pollingPeriod)

	for _, bucket := range srv.buckets {
		logger.Tracef("Starting to poll bucket %q", bucket)

		err := srv.s3Client.PollOnce(ctx, bucket, srv.s3Chan)
		if err != nil {
			return err //nolint:wrapcheck
		}

		go func(bucket string) {
			for {
				select {
				case <-time.After(pollingPeriod):
					err := srv.s3Client.PollOnce(ctx, bucket, srv.s3Chan)
					if err != nil {
						logger.Errorf("Failed to poll bucket %q: %v", bucket, err)
					}
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

		err = srv.s3Client.PollOnce(ctx, bucket, srv.s3Chan)
		if err != nil {
			return err //nolint:wrapcheck
		}
	}

	return nil
}
