package s3down

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

type AliyunCDNAuthMode string

const (
	// AliyunCDNAuthModeNone 阿里云CDN 无鉴权
	AliyunCDNAuthModeNone = ""

	// AliyunCDNAuthModeA 阿里云CDN 鉴权方式A
	// 参考: https://help.Aliyun.com/zh/cdn/user-guide/type-a-signing
	AliyunCDNAuthModeA = "type-a"

	// AliyunCDNAuthModeB 阿里云CDN 鉴权方式B
	// 参考: https://help.Aliyun.com/zh/cdn/user-guide/type-b-signing
	AliyunCDNAuthModeB = "type-b"

	// AliyunCDNAuthModeC 阿里云CDN 鉴权方式C
	// 参考: https://help.Aliyun.com/zh/cdn/user-guide/type-c-signing
	AliyunCDNAuthModeC = "type-c"

	// AliyunCDNAuthModeF 阿里云CDN 鉴权方式F
	// 参考: https://help.Aliyun.com/zh/cdn/user-guide/authentication-method-f-description
	//
	// 控制台配置要求
	// - 签名参数：sign
	// - 时间戳参数：time
	// - 时间戳格式：十六进制（Unix 时间戳）
	// - URL编码：关闭
	AliyunCDNAuthModeF = "type-f"
)

type GeneratorAliyunCDNConfig struct {
	GeneratorConfigCommon

	// Endpoint 填写CDN URL，例如：https://cdn.example.com
	Endpoint string `json:"endpoint"`

	// AuthMode 填写控制台里的 “鉴权模式”
	AuthMode AliyunCDNAuthMode `json:"auth_mode"`

	// AuthKey 填写控制台里的 “主KEY” 或 “副KEY”
	AuthKey string `json:"auth_key"`

	// DynamicExpire 生成的签名直接使用国企时间作为时间戳 (timestamp = ExpiredAt)
	// 因此开启后，在控制台必须将 “鉴权URL有效时长” 设置为 0
	DynamicExpire bool `json:"dynamic_expire"`
}

type GeneratorAliyunCDN struct {
	endpoint *url.URL
	cfg      *GeneratorAliyunCDNConfig
}

func NewGeneratorAliyunCDN(cfg *GeneratorAliyunCDNConfig) (*GeneratorAliyunCDN, error) {
	u, err := url.Parse(cfg.Endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to parse endpoint: %w", err)
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, errors.New("endpoint scheme must be http or https")
	}

	if cfg.AuthMode != AliyunCDNAuthModeNone && cfg.AuthKey == "" {
		return nil, errors.New("auth key is required")
	}

	return &GeneratorAliyunCDN{
		cfg:      cfg,
		endpoint: u,
	}, nil

}

func (d *GeneratorAliyunCDN) GenerateDownload(_ context.Context, params *GenerateParams) (*url.URL, error) {
	query := make(url.Values)

	if params.ContentType != "" {
		query.Set("response-content-type", params.ContentType)
	}

	if params.AttachmentFilename != "" {
		query.Set("response-content-disposition", composeContentDisposition(params.AttachmentFilename))
	}

	u := composeObjectURL(d.endpoint, d.cfg.Prefix, params.RemotePath)

	switch d.cfg.AuthMode {
	case AliyunCDNAuthModeA:
		d.signModeA(u, query, params.ExpireIn)
	case AliyunCDNAuthModeB:
		d.signModeB(u, params.ExpireIn)
	case AliyunCDNAuthModeC:
		d.signModeC(u, params.ExpireIn)
	case AliyunCDNAuthModeF:
		d.signModeF(u, query, params.ExpireIn)
	default:
		// no-op
	}

	u.RawQuery = query.Encode()
	return u, nil
}

func (d *GeneratorAliyunCDN) signModeA(u *url.URL, query url.Values, expire time.Duration) {
	signAt := time.Now()
	if d.cfg.DynamicExpire {
		signAt = signAt.Add(expire)
	}

	ts := signAt.Unix()

	nonce := strings.ReplaceAll(uuid.NewString(), "-", "")

	signText := fmt.Sprintf("%s-%d-%s-0-%s", u.EscapedPath(), ts, nonce, d.cfg.AuthKey)
	sign := md5.Sum([]byte(signText))
	signHex := hex.EncodeToString(sign[:])

	query.Set("auth_key", fmt.Sprintf("%d-%s-0-%s", ts, nonce, signHex))
}

func (d *GeneratorAliyunCDN) signModeB(u *url.URL, expire time.Duration) {
	signAt := time.Now()
	if d.cfg.DynamicExpire {
		signAt = signAt.Add(expire)
	}

	// YYYYMMDDHHMM
	ts := signAt.In(TimezoneCST).Format("200601021504")

	signText := strings.Join([]string{d.cfg.AuthKey, ts, u.EscapedPath()}, "")
	sign := md5.Sum([]byte(signText))
	signHex := hex.EncodeToString(sign[:])

	u.Path = path.Join("/", ts, signHex, u.Path)
}

func (d *GeneratorAliyunCDN) signModeC(u *url.URL, expire time.Duration) {
	signAt := time.Now()
	if d.cfg.DynamicExpire {
		signAt = signAt.Add(expire)
	}

	ts := strconv.FormatInt(signAt.Unix(), 16)

	signText := strings.Join([]string{d.cfg.AuthKey, u.EscapedPath(), ts}, "")
	sign := md5.Sum([]byte(signText))
	signHex := hex.EncodeToString(sign[:])

	u.Path = path.Join("/", signHex, ts, u.Path)
}

func (d *GeneratorAliyunCDN) signModeF(u *url.URL, query url.Values, expire time.Duration) {
	signAt := time.Now()
	if d.cfg.DynamicExpire {
		signAt = signAt.Add(expire)
	}

	ts := strconv.FormatInt(signAt.Unix(), 16)

	signText := strings.Join([]string{d.cfg.AuthKey, u.EscapedPath(), ts}, "")
	sign := md5.Sum([]byte(signText))
	signHex := hex.EncodeToString(sign[:])

	u.Path = path.Join("/", signHex, ts, u.Path)

	query.Set("sign", signHex)
	query.Set("time", ts)
}
