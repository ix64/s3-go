package s3up

import (
	"context"
	"net/url"
	"time"
)

type GenerateParams struct {
	// required, path of pending file
	RemotePath string

	// required, pre-signed url expiration
	ExpireIn time.Duration

	// required, size of pending file
	Size int64

	// optional, content-type of pending file
	ContentType string

	// optional, sha256 checksum of pending file
	Sha256 []byte

	// optional, object metadata
	Metadata map[string]string
}

type GenerateResult struct {
	URL      *url.URL
	FormData map[string]string
}

// Generator 为终端用户生成预签名的下载链接，一般由对象存储或CDN服务提供
type Generator interface {
	GenerateUpload(ctx context.Context, params *GenerateParams) (*GenerateResult, error)
}
