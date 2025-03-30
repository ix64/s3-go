package s3down

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"path"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

type TencentCloudCDNAuthMode string

const (
	// TencentCloudCDNAuthModeNone 腾讯云CDN 无鉴权
	TencentCloudCDNAuthModeNone = ""

	// TencentCloudCDNAuthModeA 腾讯云CDN 鉴权方式A
	// 参考: https://cloud.tencent.com/document/product/228/41623
	//
	// 控制台配置要求
	// - 鉴权算法：sha256
	// - 签名参数：sign
	TencentCloudCDNAuthModeA = "type-a"

	// TencentCloudCDNAuthModeB 腾讯云CDN 鉴权方式B
	// 参考: https://cloud.tencent.com/document/product/228/41871
	//
	// 控制台配置要求
	// - 鉴权算法：sha256
	TencentCloudCDNAuthModeB = "type-b"

	// TencentCloudCDNAuthModeC 腾讯云CDN 鉴权方式C
	// 参考: https://cloud.tencent.com/document/product/228/41624
	//
	// 控制台配置要求
	// - 鉴权算法：sha256
	TencentCloudCDNAuthModeC = "type-c"

	// TencentCloudCDNAuthModeD 腾讯云CDN 鉴权方式F
	// 参考: https://cloud.tencent.com/document/product/228/41625
	//
	// 控制台配置要求
	// - 鉴权算法：sha256
	// - 签名参数：sign
	// - 时间戳参数：t
	// - 时间戳格式：十六进制（Unix 时间戳）
	TencentCloudCDNAuthModeD = "type-d"
)

var TencentCloudCDNAuthModes = []TencentCloudCDNAuthMode{
	TencentCloudCDNAuthModeNone,
	TencentCloudCDNAuthModeA,
	TencentCloudCDNAuthModeB,
	TencentCloudCDNAuthModeC,
	TencentCloudCDNAuthModeD,
}

type GeneratorTencentCloudCDNConfig struct {
	GeneratorConfigCommon

	// Endpoint 填写CDN URL，例如：https://cdn.example.com
	Endpoint string `json:"endpoint"`

	// AuthMode 填写控制台里的 “鉴权模式”
	AuthMode TencentCloudCDNAuthMode `json:"auth_mode"`

	// AuthKey 填写控制台里的 “主KEY” 或 “副KEY”
	AuthKey string `json:"auth_key"`

	// DynamicExpire 生成的签名直接使用国企时间作为时间戳 (timestamp = ExpiredAt)
	// 因此开启后，在控制台必须将 “鉴权URL有效时长” 设置为 0
	DynamicExpire bool `json:"dynamic_expire"`
}

func (c *GeneratorTencentCloudCDNConfig) Validate() error {
	if c.Endpoint == "" {
		return errors.New("endpoint is required")
	}

	if !slices.Contains(TencentCloudCDNAuthModes, c.AuthMode) {
		return fmt.Errorf("unknown auth mode: %s", c.AuthMode)
	}

	if c.AuthMode != TencentCloudCDNAuthModeNone && c.AuthKey == "" {
		return errors.New("auth key is required")
	}

	return nil
}

type GeneratorTencentCloudCDN struct {
	endpoint *url.URL
	cfg      *GeneratorTencentCloudCDNConfig
}

func NewGeneratorTencentCloudCDN(cfg *GeneratorTencentCloudCDNConfig) (*GeneratorTencentCloudCDN, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	u, err := url.Parse(cfg.Endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to parse endpoint: %w", err)
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, errors.New("endpoint scheme must be http or https")
	}

	if cfg.AuthMode != TencentCloudCDNAuthModeNone && cfg.AuthKey == "" {
		return nil, errors.New("auth key is required")
	}

	return &GeneratorTencentCloudCDN{
		cfg:      cfg,
		endpoint: u,
	}, nil

}

func (d *GeneratorTencentCloudCDN) GenerateDownload(_ context.Context, params *GenerateParams) (*url.URL, error) {
	query := make(url.Values)

	if !d.cfg.DisableResponseContentType && params.ContentType != "" {
		query.Set("response-content-type", params.ContentType)
	}

	if !d.cfg.DisableResponseContentDisposition && params.AttachmentFilename != "" {
		query.Set("response-content-disposition", composeContentDisposition(params.AttachmentFilename))
	}

	u := composeObjectURL(d.endpoint, d.cfg.Prefix, params.RemotePath)

	switch d.cfg.AuthMode {
	case TencentCloudCDNAuthModeA:
		d.signModeA(u, query, params.ExpireIn)
	case TencentCloudCDNAuthModeB:
		d.signModeB(u, params.ExpireIn)
	case TencentCloudCDNAuthModeC:
		d.signModeC(u, params.ExpireIn)
	case TencentCloudCDNAuthModeD:
		d.signModeD(u, query, params.ExpireIn)
	default:
		// no-op
	}

	u.RawQuery = query.Encode()
	return u, nil
}

func (d *GeneratorTencentCloudCDN) signModeA(u *url.URL, query url.Values, expire time.Duration) {
	signAt := time.Now()
	if d.cfg.DynamicExpire {
		signAt = signAt.Add(expire)
	}

	ts := signAt.Unix()

	nonce := strings.ReplaceAll(uuid.NewString(), "-", "")

	signText := fmt.Sprintf("%s-%d-%s-0-%s", u.EscapedPath(), ts, nonce, d.cfg.AuthKey)
	sign := sha256.Sum256([]byte(signText))
	signHex := hex.EncodeToString(sign[:])

	query.Set("sign", fmt.Sprintf("%d-%s-0-%s", ts, nonce, signHex))
}

func (d *GeneratorTencentCloudCDN) signModeB(u *url.URL, expire time.Duration) {
	signAt := time.Now()
	if d.cfg.DynamicExpire {
		signAt = signAt.Add(expire)
	}

	// YYYYMMDDHHMM
	ts := signAt.In(TimezoneCST).Format("200601021504")

	signText := strings.Join([]string{d.cfg.AuthKey, ts, u.EscapedPath()}, "")
	sign := sha256.Sum256([]byte(signText))
	signHex := hex.EncodeToString(sign[:])

	u.Path = path.Join("/", ts, signHex, u.Path)
}

func (d *GeneratorTencentCloudCDN) signModeC(u *url.URL, expire time.Duration) {
	signAt := time.Now()
	if d.cfg.DynamicExpire {
		signAt = signAt.Add(expire)
	}

	ts := strconv.FormatInt(signAt.Unix(), 16)

	signText := strings.Join([]string{d.cfg.AuthKey, u.EscapedPath(), ts}, "")
	sign := sha256.Sum256([]byte(signText))
	signHex := hex.EncodeToString(sign[:])

	u.Path = path.Join("/", signHex, ts, u.Path)
}

func (d *GeneratorTencentCloudCDN) signModeD(u *url.URL, query url.Values, expire time.Duration) {
	signAt := time.Now()
	if d.cfg.DynamicExpire {
		signAt = signAt.Add(expire)
	}

	ts := strconv.FormatInt(signAt.Unix(), 16)

	signText := strings.Join([]string{d.cfg.AuthKey, u.EscapedPath(), ts}, "")
	sign := sha256.Sum256([]byte(signText))
	signHex := hex.EncodeToString(sign[:])

	query.Set("sign", signHex)
	query.Set("t", ts)
}
