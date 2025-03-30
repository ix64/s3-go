package s3

import (
	"encoding/json"
	"fmt"

	"github.com/ix64/s3-go/s3up"
)

type UploadGeneratorType string

const (
	UploadGeneratorTypeS3 UploadGeneratorType = "s3"
)

func newUploadGenerator(c *Client, t UploadGeneratorType, raw json.RawMessage) (s3up.Generator, error) {
	switch t {
	case UploadGeneratorTypeS3:
		cfg := &s3up.GeneratorS3Config{}
		if raw != nil {
			if err := json.Unmarshal(raw, cfg); err != nil {
				return nil, fmt.Errorf("failed to unmarshal config: %w", err)
			}
		}
		fillUploadGeneratorS3Defaults(cfg, c)
		return s3up.NewGeneratorS3(cfg)
	default:
		return nil, fmt.Errorf("unknown Upload Generator type: %s", t)
	}
}

func fillUploadGeneratorS3Defaults(cfg *s3up.GeneratorS3Config, c *Client) {
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
