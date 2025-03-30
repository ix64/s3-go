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

	// DisableResponseContentType 部分S3供应商设置 "response-content-type" header 会导致报错。
	// 由于 CDN 供应商可能会将此 header 回源到 S3 服务，所有实现都应该参考此配置
	//   已知不支持的供应商
	//   - 阿里云 OSS: https://help.aliyun.com/zh/oss/support/0017-00000902
	DisableResponseContentType bool `json:"disable_response_content_type"`
}
