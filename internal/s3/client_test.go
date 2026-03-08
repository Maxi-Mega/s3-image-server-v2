package s3

import (
	"reflect"
	"testing"

	"github.com/Maxi-Mega/s3-image-server-v2/config"
)

func TestNewClient(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name                         string
		cfg                          config.Config
		expectedSpecificInfo         map[string]bucketSpecificInfo
		expectedCommonPrefixByBucket map[string]string
	}{
		{
			name: "single bucket multiple product prefixes",
			cfg: config.Config{
				S3: config.S3{
					Endpoint: "localhost:9000",
				},
				Products: config.Products{
					ImageGroups: []config.ImageGroup{
						{
							Bucket: "bucket-a",
							Types: []config.ImageType{
								{ProductPrefix: "products/a/preview/"},
								{ProductPrefix: "products/a/target/"},
							},
						},
					},
				},
			},
			expectedSpecificInfo: map[string]bucketSpecificInfo{
				"bucket-a": {commonPrefix: "products/a/"},
			},
			expectedCommonPrefixByBucket: map[string]string{
				"bucket-a": "products/a/",
			},
		},
		{
			name: "shared bucket across groups and independent bucket",
			cfg: config.Config{
				S3: config.S3{
					Endpoint: "localhost:9000",
				},
				Products: config.Products{
					ImageGroups: []config.ImageGroup{
						{
							Bucket: "bucket-shared",
							Types: []config.ImageType{
								{ProductPrefix: "root/a/type1/"},
							},
						},
						{
							Bucket: "bucket-shared",
							Types: []config.ImageType{
								{ProductPrefix: "root/a/type2/"},
								{ProductPrefix: "root/b/type3/"},
							},
						},
						{
							Bucket: "bucket-other",
							Types: []config.ImageType{
								{ProductPrefix: "other/prefix/"},
							},
						},
					},
				},
			},
			expectedSpecificInfo: map[string]bucketSpecificInfo{
				"bucket-shared": {commonPrefix: "root/"},
				"bucket-other":  {commonPrefix: "other/prefix/"},
			},
			expectedCommonPrefixByBucket: map[string]string{
				"bucket-shared": "root/",
				"bucket-other":  "other/prefix/",
			},
		},
		{
			name: "bucket with one type and bucket without types is ignored",
			cfg: config.Config{
				S3: config.S3{
					Endpoint: "localhost:9000",
				},
				Products: config.Products{
					ImageGroups: []config.ImageGroup{
						{
							Bucket: "bucket-one",
							Types: []config.ImageType{
								{ProductPrefix: "only-prefix"},
							},
						},
						{
							Bucket: "bucket-empty",
							Types:  []config.ImageType{},
						},
					},
				},
			},
			expectedSpecificInfo: map[string]bucketSpecificInfo{
				"bucket-one": {commonPrefix: "only-prefix"},
			},
			expectedCommonPrefixByBucket: map[string]string{
				"bucket-one": "only-prefix",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			client, err := NewClient(tc.cfg, nil)
			if err != nil {
				t.Fatalf("NewClient() error = %v", err)
			}

			typedClient, ok := client.(s3Client)
			if !ok {
				t.Fatalf("NewClient() returned unexpected type %T", client)
			}

			if !reflect.DeepEqual(typedClient.specificInfoPerBucket, tc.expectedSpecificInfo) {
				t.Fatalf("specificInfoPerBucket mismatch: got %#v, want %#v", typedClient.specificInfoPerBucket, tc.expectedSpecificInfo)
			}

			if !reflect.DeepEqual(typedClient.commonPrefixesPerBucket, tc.expectedCommonPrefixByBucket) {
				t.Fatalf("commonPrefixesPerBucket mismatch: got %#v, want %#v", typedClient.commonPrefixesPerBucket, tc.expectedCommonPrefixByBucket)
			}
		})
	}
}
