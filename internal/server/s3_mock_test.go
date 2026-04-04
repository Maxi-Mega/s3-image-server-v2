package server

import (
	"context"
	"net/url"
	"time"

	"github.com/Maxi-Mega/s3-image-server-v2/internal/s3"
)

type S3ClientMock struct {
	BucketExistsFn      func(ctx context.Context, bucket string) (bool, error)
	SubscribeToBucketFn func(ctx context.Context, bucket string, s3Chan chan s3.Event) error
	PollOnceFn          func(ctx context.Context, bucket string, s3Chan chan s3.Event, timeout time.Duration) error
	DownloadObjectFn    func(ctx context.Context, bucket, objectKey, destPath string) error
	GenerateSignedURLFn func(ctx context.Context, bucket, objectKey string) (*url.URL, error)
}

func (s3 S3ClientMock) BucketExists(ctx context.Context, bucket string) (bool, error) {
	return s3.BucketExistsFn(ctx, bucket)
}

func (s3 S3ClientMock) SubscribeToBucket(ctx context.Context, bucket string, s3Chan chan s3.Event) error {
	return s3.SubscribeToBucketFn(ctx, bucket, s3Chan)
}

func (s3 S3ClientMock) PollOnce(ctx context.Context, bucket string, s3Chan chan s3.Event, timeout time.Duration) error {
	return s3.PollOnceFn(ctx, bucket, s3Chan, timeout)
}

func (s3 S3ClientMock) DownloadObject(ctx context.Context, bucket, objectKey, destPath string) error {
	return s3.DownloadObjectFn(ctx, bucket, objectKey, destPath)
}

func (s3 S3ClientMock) GenerateSignedURL(ctx context.Context, bucket, objectKey string) (*url.URL, error) {
	return s3.GenerateSignedURLFn(ctx, bucket, objectKey)
}
