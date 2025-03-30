package s3up

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"github.com/ix64/s3-go/s3type"
)

type GeneratorS3Config struct {
	Endpoint     string                  `json:"endpoint"`
	Bucket       string                  `json:"bucket"`
	BucketLookup s3type.BucketLookupType `json:"bucket_lookup"`
	Prefix       string                  `json:"prefix"`

	Region string `json:"region"`

	AccessKey string `json:"accessKey"`
	SecretKey string `json:"secretKey"`

	ChecksumEnabled bool `json:"checksum_enabled"` // enable sha256 checksum
}
type GeneratorS3 struct {
	client *minio.Client
	cfg    *GeneratorS3Config
}

func checkGeneratorS3Config(cfg *GeneratorS3Config) error {
	if cfg.Endpoint == "" {
		return errors.New("endpoint is required")
	}

	if cfg.Bucket == "" {
		return errors.New("bucket is required")
	}

	if cfg.Region == "" {
		return errors.New("region is required")
	}

	if cfg.AccessKey == "" || cfg.SecretKey == "" {
		return errors.New("access key and secret key is required when anonymous is false")
	}

	return nil
}

func NewGeneratorS3(cfg *GeneratorS3Config) (*GeneratorS3, error) {
	if err := checkGeneratorS3Config(cfg); err != nil {
		return nil, err
	}

	u, err := url.Parse(cfg.Endpoint)
	if err != nil {
		return nil, err
	}

	var bucketLookup minio.BucketLookupType
	switch cfg.BucketLookup {
	case s3type.BucketLookupDNS:
		bucketLookup = minio.BucketLookupDNS
	case s3type.BucketLookupPath:
		bucketLookup = minio.BucketLookupPath
	case s3type.BucketLookupCNAME:
		return nil, errors.New("custom domain by CNAME is not supported for S3 UploadGenerator")
	default:
		return nil, fmt.Errorf("unknown bucket lookup type: %s", cfg.BucketLookup)
	}

	client, err := minio.New(u.Host, &minio.Options{
		Creds:        credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure:       u.Scheme == "https",
		BucketLookup: bucketLookup,
		Region:       cfg.Region,
	})
	if err != nil {
		return nil, err
	}

	return &GeneratorS3{
		cfg:    cfg,
		client: client,
	}, nil
}

func (p *GeneratorS3) GenerateUpload(ctx context.Context, params *GenerateParams) (*GenerateResult, error) {
	policy := minio.NewPostPolicy()

	// enforce bucket name
	if err := policy.SetBucket(p.cfg.Bucket); err != nil {
		return nil, err
	}

	// enforce object name
	objectName := composeObjectName(p.cfg.Prefix, params.RemotePath)
	if err := policy.SetKey(objectName); err != nil {
		return nil, err
	}

	// enforce file size
	if err := policy.SetContentLengthRange(params.Size, params.Size); err != nil {
		return nil, err
	}

	// set link expire
	expireAt := time.Now().Add(params.ExpireIn)
	if err := policy.SetExpires(expireAt); err != nil {
		return nil, err
	}

	// optionally enforce content type
	if params.ContentType != "" {
		if err := policy.SetContentType(params.ContentType); err != nil {
			return nil, err
		}
	}

	// optionally enforce file sha256 checksum
	if p.cfg.ChecksumEnabled && params.Sha256 != nil {
		checksum := minio.NewChecksum(minio.ChecksumSHA256, params.Sha256)

		if err := policy.SetChecksum(checksum); err != nil {
			return nil, err
		}
	}

	if params.Metadata != nil {
		for k, v := range params.Metadata {
			if err := policy.SetUserMetadata(k, v); err != nil {
				return nil, err
			}
		}
	}

	u, formData, err := p.client.PresignedPostPolicy(ctx, policy)
	if err != nil {
		return nil, err
	}

	return &GenerateResult{
		URL:      u,
		FormData: formData,
	}, nil

}
