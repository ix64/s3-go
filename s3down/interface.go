package s3down

import (
	"context"
	"net/url"
	"time"
)

type GenerateParams struct {
	RemotePath string
	ExpireIn   time.Duration

	// optional, expect to response with "Content-Disposition" header
	AttachmentFilename string

	// optional, expect to response with "Content-Type" header
	ContentType string
}

// Generator 为终端用户生成预签名的下载链接，一般由对象存储或CDN服务提供
type Generator interface {
	GenerateDownload(ctx context.Context, params *GenerateParams) (*url.URL, error)
}

type GeneratorConfigCommon struct {
	Prefix string `json:"prefix"`
}
