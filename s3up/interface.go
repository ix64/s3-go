package s3up

import (
	"context"
	"net/http"
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

	// optional, attachment filename while downloading
	AttachmentFilename string

	// optional, sha256 checksum of pending file
	Sha256 []byte

	// optional, object metadata
	Metadata map[string]string
}

type GenerateResult struct {
	Method   string
	URL      *url.URL
	Header   http.Header
	FormData map[string]string
}

// Generator 为终端用户生成预签名的下载链接，一般由对象存储或CDN服务提供
type Generator interface {
	GenerateUpload(ctx context.Context, params *GenerateParams) (*GenerateResult, error)
}
