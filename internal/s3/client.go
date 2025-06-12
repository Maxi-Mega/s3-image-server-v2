package s3

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/Maxi-Mega/s3-image-server-v2/config"
	"github.com/Maxi-Mega/s3-image-server-v2/internal/logger"
	"github.com/Maxi-Mega/s3-image-server-v2/internal/types"
	"github.com/Maxi-Mega/s3-image-server-v2/utils"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/minio/minio-go/v7/pkg/notification"
)

const (
	healthCheckTimeout    = 5 * time.Second
	pollBucketTimeout     = 2 * time.Minute
	downloadObjectTimeout = 30 * time.Second
	SignedURLLifetime     = 7 * 24 * time.Hour
)

var (
	errChanClosed = errors.New("the channel is closed")
)

type Client interface {
	BucketExists(ctx context.Context, bucket string) (bool, error)
	SubscribeToBucket(ctx context.Context, bucket string, s3Chan chan Event) error
	PollOnce(ctx context.Context, bucket string, s3Chan chan Event) error
	DownloadObject(ctx context.Context, bucket, objectKey, destPath string) error
	GenerateSignedURL(ctx context.Context, bucket, objectKey string) (string, error)
}

type s3Client struct {
	productsCfg           config.Products
	specificInfoPerBucket map[string]bucketSpecificInfo
	client                *minio.Client
}

type bucketSpecificInfo struct {
	commonPrefix              string
	notDefaultPreviewSuffixes []string
}

func NewClient(cfg config.Config) (Client, error) {
	client, err := minio.New(cfg.S3.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.S3.AccessID, cfg.S3.AccessSecret, ""),
		Secure: cfg.S3.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize s3 client: %w", err)
	}

	prefixesPerBucket := make(map[string][]string, len(cfg.Products.ImageGroups))
	previewSuffixesPerBucket := make(map[string][]string)

	for _, imgGroup := range cfg.Products.ImageGroups {
		for _, imgType := range imgGroup.Types {
			prefixesPerBucket[imgGroup.Bucket] = append(prefixesPerBucket[imgGroup.Bucket], imgType.ProductPrefix)

			if imgType.PreviewSuffix != "" {
				previewSuffixesPerBucket[imgGroup.Bucket] = append(previewSuffixesPerBucket[imgGroup.Bucket], imgType.PreviewSuffix)
			}
		}
	}

	commonPrefixPerBucket := make(map[string]bucketSpecificInfo, len(prefixesPerBucket))

	for bucket, prefixes := range prefixesPerBucket {
		commonPrefixPerBucket[bucket] = bucketSpecificInfo{
			commonPrefix:              utils.CommonPrefix(prefixes...),
			notDefaultPreviewSuffixes: previewSuffixesPerBucket[bucket],
		}
	}

	return s3Client{
		productsCfg:           cfg.Products,
		specificInfoPerBucket: commonPrefixPerBucket,
		client:                client,
	}, nil
}

func (s3 s3Client) BucketExists(ctx context.Context, bucket string) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, healthCheckTimeout)
	defer cancel()

	exists, err := s3.client.BucketExists(ctx, bucket)
	if err != nil {
		return false, fmt.Errorf("can't check for existence of bucket %q: %w", bucket, err)
	}

	return exists, nil
}

func (s3 s3Client) SubscribeToBucket(ctx context.Context, bucket string, s3Chan chan Event) error {
	var (
		previewEvents      = []string{"s3:ObjectCreated:*", "s3:ObjectRemoved:*"}
		geonameEvents      = []string{"s3:ObjectCreated:*"}
		localizationEvents = []string{"s3:ObjectCreated:*"}
		fullProductEvents  = []string{"s3:ObjectCreated:*"}
	)

	specificInfo := s3.specificInfoPerBucket[bucket]

	previewNotifs := s3.client.ListenBucketNotification(ctx, bucket, specificInfo.commonPrefix, s3.productsCfg.DefaultPreviewSuffix, previewEvents)
	geonameNotifs := s3.client.ListenBucketNotification(ctx, bucket, specificInfo.commonPrefix, s3.productsCfg.GeonamesFilename, geonameEvents)
	localizationNotifs := s3.client.ListenBucketNotification(ctx, bucket, specificInfo.commonPrefix, s3.productsCfg.LocalizationFilename, localizationEvents)
	fullProductNotifs := s3.client.ListenBucketNotification(ctx, bucket, specificInfo.commonPrefix, s3.productsCfg.FullProductExtension, fullProductEvents)

	time.Sleep(10 * time.Millisecond) // Let the time for errors to occur

	err := ensureNoError(previewNotifs, geonameNotifs, localizationNotifs, fullProductNotifs)
	if err != nil {
		return fmt.Errorf("failed to subscribe to notifications of bucket %q: %w", bucket, err)
	}

	go func() {
		logger.Debugf("Starting to listen for notifications from bucket %q", bucket)

		for {
			select {
			case notif := <-previewNotifs:
				s3.handleEvent(bucket, types.ObjectPreview, notif, s3Chan)
			case notif := <-geonameNotifs:
				s3.handleEvent(bucket, types.ObjectGeonames, notif, s3Chan)
			case notif := <-localizationNotifs:
				s3.handleEvent(bucket, types.ObjectLocalization, notif, s3Chan)
			case notif := <-fullProductNotifs:
				s3.handleEvent(bucket, types.ObjectFullProduct, notif, s3Chan)
			case <-ctx.Done():
				logger.Info("Context expired, closing event channel")

				close(s3Chan)

				return
			}
		}
	}()

	return nil
}

func (s3 s3Client) PollOnce(ctx context.Context, bucket string, s3Chan chan Event) error {
	logger.Debugf("Polling bucket %q ...", bucket)

	ctx, cancel := context.WithTimeout(ctx, pollBucketTimeout)
	defer cancel()

	commonPrefix := s3.specificInfoPerBucket[bucket].commonPrefix
	currentTime := time.Now()

	objects := s3.client.ListObjects(ctx, bucket, minio.ListObjectsOptions{Prefix: commonPrefix, Recursive: true})
	for object := range objects {
		event := Event{
			Time:               currentTime,
			Bucket:             bucket,
			EventType:          types.EventCreated,
			ObjectType:         "", // We don't know yet
			Size:               object.Size,
			ObjectKey:          object.Key,
			ObjectLastModified: object.LastModified,
		}

		s3Chan <- event
	}

	return nil
}

func (s3 s3Client) DownloadObject(ctx context.Context, bucket, objectKey, destPath string) error {
	ctx, cancel := context.WithTimeout(ctx, downloadObjectTimeout)
	defer cancel()

	return s3.client.FGetObject(ctx, bucket, objectKey, destPath, minio.GetObjectOptions{}) //nolint:wrapcheck
}

func (s3 s3Client) GenerateSignedURL(ctx context.Context, bucket, objectKey string) (string, error) {
	signedURL, err := s3.client.PresignedGetObject(ctx, bucket, objectKey, SignedURLLifetime, url.Values{})
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL for object %q in bucket %q: %w", objectKey, bucket, err)
	}

	withoutSchemeAndHost := strings.TrimPrefix(signedURL.String(), signedURL.Scheme+"://"+signedURL.Host)

	return s3.productsCfg.FullProductProtocol + url.QueryEscape(s3.productsCfg.FullProductRootURL+withoutSchemeAndHost), nil
}

func (s3 s3Client) handleEvent(bucket string, objectType types.ObjectType, notif notification.Info, eventChan chan Event) {
	if notif.Err != nil {
		logger.Errorf("Received error from bucket %q: %v", bucket, notif.Err)

		return
	}

	logger.Debugf("Handling S3 event from bucket %q on %s object", bucket, objectType)

	for _, e := range notif.Records {
		if !s3.matchesPreviewFilename(bucket, e.S3.Object.Key) {
			continue
		}

		eventTime := parseEventTime(e.EventTime)
		event := Event{
			Time:               eventTime,
			Bucket:             e.S3.Bucket.Name,
			EventType:          parseEventType(e.EventName),
			ObjectType:         objectType,
			Size:               e.S3.Object.Size,
			ObjectKey:          e.S3.Object.Key,
			ObjectLastModified: eventTime,
		}

		eventChan <- event
	}
}

// matchesPreviewFilename ensure that the given object's key matches the expected preview filename.
// It may be whether the defaultPreviewFilename or the image type's specific previewFilename,
// since a more granular filter will be done later.
func (s3 s3Client) matchesPreviewFilename(bucket, objectKey string) bool {
	specificInfo, bucketExists := s3.specificInfoPerBucket[bucket]
	if !bucketExists {
		return false
	}

	if s3.productsCfg.DefaultPreviewSuffix != "" && strings.HasSuffix(objectKey, s3.productsCfg.DefaultPreviewSuffix) {
		return true
	}

	for _, previewSuffix := range specificInfo.notDefaultPreviewSuffixes {
		if strings.HasSuffix(objectKey, previewSuffix) {
			return true
		}
	}

	return false
}

// ensureNoError tries to read from each given channel
// to check whether an error has been sent or not.
func ensureNoError(channels ...<-chan notification.Info) error {
	for _, c := range channels {
		select {
		case notif, ok := <-c:
			if !ok {
				return errChanClosed
			}

			if notif.Err != nil {
				return notif.Err
			}
		default:
			continue
		}
	}

	return nil
}

func parseEventTime(rawTime string) time.Time {
	const eventTimeLayout = "2006-01-02T15:04:05.000Z"

	eventTime, err := time.Parse(eventTimeLayout, rawTime)
	if err != nil {
		logger.Errorf("Failed to parse event time %q", rawTime)

		return time.Now() // probably not too far from reality
	}

	return eventTime
}
