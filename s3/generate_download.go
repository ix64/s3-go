package s3

import (
	"encoding/json"
	"fmt"

	s3down2 "github.com/ix64/s3-go/s3down"
)

type DownloadGeneratorType string

const (
	DownloadGeneratorTypeS3              DownloadGeneratorType = "s3type"
	DownloadGeneratorTypeAliyunCDN       DownloadGeneratorType = "aliyun_cdn"
	DownloadGeneratorTypeTencentCloudCDN DownloadGeneratorType = "tencent_cloud_cdn"
)

func newDownloadGenerator(c *Client, t DownloadGeneratorType, raw json.RawMessage) (s3down2.Generator, error) {
	switch t {
	case DownloadGeneratorTypeS3, "": // default
		cfg := &s3down2.GeneratorS3Config{}
		if err := json.Unmarshal(raw, cfg); err != nil {
			return nil, fmt.Errorf("failed to unmarshal config: %w", err)
		}
		fillDownloadGeneratorS3Defaults(cfg, c)
		return s3down2.NewGeneratorS3(cfg)

	case DownloadGeneratorTypeAliyunCDN:
		cfg := &s3down2.GeneratorAliyunCDNConfig{}
		if err := json.Unmarshal(raw, cfg); err != nil {
			return nil, fmt.Errorf("failed to unmarshal config: %w", err)
		}
		return s3down2.NewGeneratorAliyunCDN(cfg)

	case DownloadGeneratorTypeTencentCloudCDN:
		cfg := &s3down2.GeneratorTencentCloudCDNConfig{}
		if err := json.Unmarshal(raw, cfg); err != nil {
			return nil, fmt.Errorf("failed to unmarshal config: %w", err)
		}
		return s3down2.NewGeneratorTencentCloudCDN(cfg)

	default:
		return nil, fmt.Errorf("unknown s3down Generator type: %s", t)
	}
}

func fillDownloadGeneratorS3Defaults(cfg *s3down2.GeneratorS3Config, c *Client) {
	if cfg.Region == "" {
		cfg.Region = c.region
	}

	if cfg.Endpoint == "" {
		cfg.Endpoint = c.cfg.Endpoint
	}
	if cfg.Bucket == "" {
		cfg.Bucket = c.cfg.Bucket
	}
	if cfg.BucketLookup == "" {
		cfg.BucketLookup = c.cfg.BucketLookup
	}

	if cfg.Prefix == "" {
		cfg.Prefix = c.prefix
	}

	if cfg.AccessKey == "" {
		cfg.AccessKey = c.cfg.AccessKey
	}

	if cfg.SecretKey == "" {
		cfg.SecretKey = c.cfg.SecretKey
	}
}
