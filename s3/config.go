package s3

import (
	"encoding/json"

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
