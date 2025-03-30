package s3

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"github.com/ix64/s3-go/s3down"
	"github.com/ix64/s3-go/s3type"
	"github.com/ix64/s3-go/s3up"
)

type Client struct {
	initialized bool

	cfg    *Config
	prefix string

	c *minio.Client

	endpoint *url.URL

	region string

	upload   s3up.Generator
	download s3down.Generator
}

// NewClient 初始化 MinIO Storage
func NewClient(cfg *Config) (c *Client, err error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("failed to validate config: %w", err)
	}

	c = &Client{
		prefix: strings.TrimPrefix(cfg.Prefix, "/"),
		cfg:    cfg,
	}

	initCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := c.init(initCtx); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Client) init(ctx context.Context) (err error) {
	if c.initialized {
		return nil
	}

	c.endpoint, err = url.Parse(c.cfg.Endpoint)
	if err != nil {
		return err
	}

	var bucketLookup minio.BucketLookupType
	switch c.cfg.BucketLookup {
	case s3type.BucketLookupDNS:
		bucketLookup = minio.BucketLookupDNS
	case s3type.BucketLookupPath:
		bucketLookup = minio.BucketLookupPath
	case s3type.BucketLookupCNAME:
		return errors.New("custom domain by CNAME is not supported for S3 UploadGenerator")
	default:
		return fmt.Errorf("unknown bucket lookup type: %s", c.cfg.BucketLookup)
	}

	c.c, err = minio.New(c.endpoint.Host, &minio.Options{
		Creds:        credentials.NewStaticV4(c.cfg.AccessKey, c.cfg.SecretKey, ""),
		Secure:       c.endpoint.Scheme == "https",
		BucketLookup: bucketLookup,
	})
	if err != nil {
		return err
	}

	c.region, err = c.c.GetBucketLocation(ctx, c.cfg.Bucket)
	if err != nil {
		return err
	}

	c.download, err = newDownloadGenerator(c, c.cfg.DownloadGeneratorType, c.cfg.DownloadGeneratorConfig)
	if err != nil {
		return fmt.Errorf("failed to init s3down Generator: %w", err)
	}

	c.upload, err = newUploadGenerator(c, c.cfg.UploadGeneratorType, c.cfg.UploadGeneratorConfig)
	if err != nil {
		return fmt.Errorf("failed to init s3up Generator: %w", err)
	}

	return nil
}

func (c *Client) composeObjectName(remotePath string) string {
	// s3 object name prefix should not start with "/"
	return strings.TrimPrefix(path.Join(c.cfg.Prefix, remotePath), "/")
}

// SetDownloadGenerator 可设置自定义的下载链接生成器
func (c *Client) SetDownloadGenerator(g s3down.Generator) {
	c.download = g
}

func (c *Client) SetUploadGenerator(g s3up.Generator) {
	c.upload = g
}
