package s3up

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"github.com/ix64/s3-go/s3common"
)

type GeneratorS3Config struct {
	Endpoint     string                    `json:"endpoint"`
	Bucket       string                    `json:"bucket"`
	BucketLookup s3common.BucketLookupType `json:"bucket_lookup"`
	Prefix       string                    `json:"prefix"`

	Region string `json:"region"`

	AccessKey string `json:"accessKey"`
	SecretKey string `json:"secretKey"`

	// DisableChecksum 部分供应商不支持 sha256 校验
	// 关闭后存在风险，即用户上传的文件 hash 值不会校验
	DisableChecksum bool `json:"checksum_enabled"`

	// DisablePOST 部分供应商不支持 Pre-signed POST 允许回退到 Pre-signed PUT
	//   已知不支持的供应商
	//   - Cloudflare R2: https://developers.cloudflare.com/r2/api/s3/presigned-urls/#supported-http-methods
	DisablePOST bool `json:"disable_post"`
}

type GeneratorS3 struct {
	client *minio.Client
	cfg    *GeneratorS3Config
}

func (c *GeneratorS3Config) Validate() error {
	if c.Endpoint == "" {
		return errors.New("endpoint is required")
	}

	if c.Bucket == "" {
		return errors.New("bucket is required")
	}

	if c.Region == "" {
		return errors.New("region is required")
	}

	if c.AccessKey == "" || c.SecretKey == "" {
		return errors.New("access key and secret key is required when anonymous is false")
	}

	if c.BucketLookup == "" {
		return errors.New("bucket lookup is required")
	}

	return nil
}

func NewGeneratorS3(cfg *GeneratorS3Config) (Generator, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	u, err := url.Parse(cfg.Endpoint)
	if err != nil {
		return nil, err
	}

	var bucketLookup minio.BucketLookupType
	switch cfg.BucketLookup {
	case s3common.BucketLookupDNS:
		bucketLookup = minio.BucketLookupDNS
	case s3common.BucketLookupPath:
		bucketLookup = minio.BucketLookupPath
	case s3common.BucketLookupCNAME:
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
	if p.cfg.DisablePOST {
		return p.generatePUT(ctx, params)
	}
	return p.generatePOST(ctx, params)
}

func (p *GeneratorS3) generatePOST(ctx context.Context, params *GenerateParams) (*GenerateResult, error) {
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

	// enforce content type
	if params.ContentType != "" {
		if err := policy.SetContentType(params.ContentType); err != nil {
			return nil, err
		}
	}

	// enforce attachment filename
	if params.AttachmentFilename != "" {
		if err := policy.SetContentDisposition(s3common.ComposeContentDisposition(params.AttachmentFilename)); err != nil {
			return nil, err
		}
	}

	// optionally enforce file sha256 checksum
	if !p.cfg.DisableChecksum && params.Sha256 != nil {
		checksum := minio.NewChecksum(minio.ChecksumSHA256, params.Sha256)

		if err := policy.SetChecksum(checksum); err != nil {
			return nil, err
		}
	}

	// set user metadata
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
		Method:   http.MethodPost,
		URL:      u,
		FormData: formData,
	}, nil

}

const (
	headerContentType          = "Content-Type"
	headerAmzChecksumAlgorithm = "x-amz-checksum-algorithm"
	headerUserMetadataPrefix   = "x-amz-meta-"
)

func (p *GeneratorS3) generatePUT(ctx context.Context, params *GenerateParams) (*GenerateResult, error) {
	header := http.Header{}

	// enforce content length
	header.Set("Content-Length", strconv.FormatInt(params.Size, 10))

	// enforce content type
	if params.ContentType == "" {
		params.ContentType = "application/octet-stream"
	}
	header.Set(headerContentType, params.ContentType)

	// optionally enforce file sha256 checksum
	if !p.cfg.DisableChecksum && params.Sha256 != nil {
		checksum := minio.NewChecksum(minio.ChecksumSHA256, params.Sha256)
		header.Set(headerAmzChecksumAlgorithm, checksum.Type.Key())
		header.Set(checksum.Type.Key(), checksum.Encoded())
	}

	// set user metadata
	if params.Metadata != nil {
		for k, v := range params.Metadata {
			header.Set(headerUserMetadataPrefix+k, v)
		}
	}

	u, err := p.client.PresignHeader(ctx,
		http.MethodPut,
		p.cfg.Bucket,
		composeObjectName(p.cfg.Prefix, params.RemotePath),
		params.ExpireIn,
		nil,
		header,
	)
	if err != nil {
		return nil, err
	}

	return &GenerateResult{
		Method: http.MethodPut,
		URL:    u,
		Header: header,
	}, nil
}
