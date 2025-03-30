package s3down

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"

	"github.com/minio/minio-go/v7/pkg/s3utils"
	"github.com/minio/minio-go/v7/pkg/signer"

	"github.com/ix64/s3-go/s3common"
)

type GeneratorS3Config struct {
	GeneratorConfigCommon

	Endpoint     string                    `json:"endpoint"`
	Bucket       string                    `json:"bucket"`
	BucketLookup s3common.BucketLookupType `json:"bucket_lookup"`

	Region string `json:"region"`

	PublicRead bool   `json:"public_read"`
	AccessKey  string `json:"access_key"`
	SecretKey  string `json:"secret_key"`
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

	if !c.PublicRead && (c.AccessKey == "" || c.SecretKey == "") {
		return errors.New("access key and secret key is required when anonymous is false")
	}

	if c.BucketLookup == "" {
		return errors.New("bucket lookup is required")
	}

	return nil
}

func NewGeneratorS3(cfg *GeneratorS3Config) (*GeneratorS3, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	u, err := url.Parse(cfg.Endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to parse endpoint: %w", err)
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, errors.New("endpoint scheme must be http or https")
	}

	switch cfg.BucketLookup {
	case s3common.BucketLookupDNS:
		u.Host = cfg.Bucket + "." + u.Host
	case s3common.BucketLookupPath:
		u.Path = path.Join("/", cfg.Bucket, u.Path)
	case s3common.BucketLookupCNAME:
		// do nothing
	default:
		return nil, fmt.Errorf("unknown bucket lookup type: %s", cfg.BucketLookup)
	}

	return &GeneratorS3{
		endpoint: u,
		cfg:      cfg,
	}, nil
}

// GeneratorS3 returns s3 pre-signed GET Object URL
type GeneratorS3 struct {
	cfg      *GeneratorS3Config
	endpoint *url.URL
}

func (d *GeneratorS3) GenerateDownload(_ context.Context, params *GenerateParams) (*url.URL, error) {
	reqParams := make(url.Values)

	if !d.cfg.DisableResponseContentType && params.ContentType != "" {
		reqParams.Set("response-content-type", params.ContentType)
	}

	if !d.cfg.DisableResponseContentDisposition && params.AttachmentFilename != "" {
		reqParams.Set("response-content-disposition", s3common.ComposeContentDisposition(params.AttachmentFilename))
	}

	ret := composeObjectURL(d.endpoint, d.cfg.Prefix, params.RemotePath)
	ret.RawPath = s3utils.EncodePath(ret.Path)
	ret.RawQuery = reqParams.Encode()

	if !d.cfg.PublicRead {
		// this request never send, skip context and error
		req, _ := http.NewRequest(http.MethodGet, ret.String(), nil)

		expireIn := int64(params.ExpireIn.Seconds())
		req = signer.PreSignV4(*req, d.cfg.AccessKey, d.cfg.SecretKey, "", d.cfg.Region, expireIn)

		ret = req.URL
	}

	return ret, nil
}
