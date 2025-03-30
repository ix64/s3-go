package s3

import (
	"encoding/json"
	"fmt"

	"github.com/ix64/s3-go/s3type"
)

type Config struct {
	Endpoint     string                  `json:"endpoint"`
	Bucket       string                  `json:"bucket"`
	BucketLookup s3type.BucketLookupType `json:"bucket_lookup"`
	Prefix       string                  `json:"prefix"`

	AccessKey string `json:"access_key"`
	SecretKey string `json:"secret_key"`

	// UploadGenerator is optional, default to s3
	UploadGeneratorType UploadGeneratorType `json:"upload_generator_type"`

	// UploadGeneratorConfig should unmarshal by Generator constructor
	UploadGeneratorConfig json.RawMessage `json:"upload_generator_config"`

	// DownloadGenerator is optional, default to s3
	DownloadGeneratorType DownloadGeneratorType `json:"download_generator_type"`

	// DownloadGeneratorConfig should unmarshal by Generator constructor
	DownloadGeneratorConfig json.RawMessage `json:"download_generator_config"`
}

func (c *Config) Validate() error {
	if c.Endpoint == "" {
		return fmt.Errorf("endpoint is required")
	}
	if c.Bucket == "" {
		return fmt.Errorf("bucket is required")
	}

	switch c.BucketLookup {
	case s3type.BucketLookupDNS,
		s3type.BucketLookupPath,
		s3type.BucketLookupCNAME:
	case "":
		return fmt.Errorf("bucket_lookup is required")
	default:
		return fmt.Errorf("unknown bucket lookup type: %s", c.BucketLookup)
	}

	if c.AccessKey == "" {
		return fmt.Errorf("access_key is required")
	}
	if c.SecretKey == "" {
		return fmt.Errorf("secret_key is required")
	}

	if c.UploadGeneratorType == "" {
		c.UploadGeneratorType = UploadGeneratorTypeS3
	}
	if c.UploadGeneratorType != UploadGeneratorTypeS3 {
		if c.UploadGeneratorConfig == nil {
			return fmt.Errorf("upload_generator_config is required")
		}
	}

	if c.DownloadGeneratorType == "" {
		c.DownloadGeneratorType = DownloadGeneratorTypeS3
	}
	if c.DownloadGeneratorType != DownloadGeneratorTypeS3 {
		if c.DownloadGeneratorConfig == nil {
			return fmt.Errorf("download_generator_config is required")
		}
	}

	return nil
}

// ParseConfig 读取 JSON 格式的配置文件
func ParseConfig(data []byte) (*Config, error) {
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("failed to validate config: %w", err)
	}

	return &cfg, nil
}
