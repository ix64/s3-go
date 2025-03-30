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
	//  已知不支持的供应商
	//  - 阿里云 OSS: 设置后报错，https://help.aliyun.com/zh/oss/support/0017-00000902
	//  - Cloudflare R2: 使用自定义域名时（公开读取）不生效，设置此选项可缩短 URL
	//
	//  对于使用CDN的用户，由于CDN的认证签名内容通常不包含Query String，此参数可能被篡改，
	//  因此建议设置此选项，并在CDN配置中禁止透传 "response-content-type"。
	//
	//  如果设置此选项后，仍希望 Response Header 中包含 Content-Type，可在 PUT Object 时设置
	DisableResponseContentType bool `json:"disable_response_content_type"`

	// DisableResponseContentDisposition 部分S3供应商设置 "response-content-disposition" header 不生效
	//   已知不支持的供应商
	//   - Cloudflare R2: 使用自定义域名时（公开读取）不生效，设置此选项可缩短 URL
	//
	//  对于使用CDN的用户，由于CDN的认证签名内容通常不包含Query String，此参数可能被篡改，
	//  因此建议设置此选项，并在CDN配置中禁止透传 "response-content-disposition"。
	//
	//  如果设置此选项后，仍希望 Response Header 中包含 Content-Disposition，可在 PUT Object 时设置
	DisableResponseContentDisposition bool `json:"disable_response_content_disposition"`
}
