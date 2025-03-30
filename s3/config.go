package s3

import (
	"encoding/json"

	"github.com/ix64/s3-go/s3type"
)

type Config struct {
	Endpoint     string `validate:"required"`
	Bucket       string `validate:"required"`
	BucketLookup s3type.BucketLookupType
	Prefix       string

	AccessKey string `validate:"required"`
	SecretKey string `validate:"required"`

	// UploadGenerator is optional, default to s3
	UploadGeneratorType UploadGeneratorType

	// UploadGeneratorConfig should unmarshal by Generator constructor
	UploadGeneratorConfig json.RawMessage

	// DownloadGenerator is optional, default to s3
	DownloadGeneratorType DownloadGeneratorType

	// DownloadGeneratorConfig should unmarshal by Generator constructor
	DownloadGeneratorConfig json.RawMessage
}
