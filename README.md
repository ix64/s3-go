# s3-go

`s3-go` 是一个面向 Go 的轻量级对象存储封装，提供两类能力：

- 服务端对象操作：上传、下载、查询、复制、移动、删除
- 面向终端用户的直传直下载链接生成：S3 预签名 URL，以及 CDN 下载鉴权 URL

当前仓库内置了：

- S3 兼容上传生成器
- S3 兼容下载生成器
- 阿里云 CDN 下载生成器
- 腾讯云 CDN 下载生成器

## 安装

```bash
go get github.com/ix64/s3-go
```

## 包结构

- `s3`：主入口，`Client`、配置解析、对象操作
- `s3up`：上传链接生成器
- `s3down`：下载链接生成器
- `s3common`：公共常量和工具函数

## 快速开始

### 1. 创建客户端

```go
package main

import (
	"context"
	"log"
	"os"

	"github.com/ix64/s3-go/s3"
)

func main() {
	cfg := &s3.Config{
		Endpoint:     "http://127.0.0.1:9000",
		Bucket:       "example-bucket",
		BucketLookup: "path",
		AccessKey:    "minio",
		SecretKey:    "E2ETestPassword",
	}

	client, err := s3.NewClient(cfg)
	if err != nil {
		log.Fatal(err)
	}

	f, err := os.Open("example.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		log.Fatal(err)
	}

	err = client.Upload(context.Background(), "demo/example.txt", f, info.Size(), "text/plain")
	if err != nil {
		log.Fatal(err)
	}
}
```

### 2. 上传和下载文件

```go
ctx := context.Background()

err := client.UploadFile(ctx, "demo/file.bin", "./file.bin")
if err != nil {
	panic(err)
}

err = client.DownloadFile(ctx, "demo/file.bin", "./downloaded.bin")
if err != nil {
	panic(err)
}
```

### 3. 读取对象信息

```go
stat, err := client.Stat(ctx, "demo/file.bin")
if err != nil {
	panic(err)
}

println(stat.Key, stat.Size)
```

### 4. 复制、移动、删除

```go
if err := client.Copy(ctx, "demo/file.bin", "demo/file-copy.bin"); err != nil {
	panic(err)
}

if err := client.Move(ctx, "demo/file-copy.bin", "archive/file.bin"); err != nil {
	panic(err)
}

if err := client.Delete(ctx, "archive/file.bin"); err != nil {
	panic(err)
}
```

## 预签名下载

```go
package main

import (
	"context"
	"log"
	"time"

	"github.com/ix64/s3-go/s3down"
)

func main() {
	u, err := client.GenerateDownload(context.Background(), &s3down.GenerateParams{
		RemotePath:         "demo/file.bin",
		ExpireIn:           time.Minute,
		ContentType:        "application/octet-stream",
		AttachmentFilename: "file.bin",
	})
	if err != nil {
		log.Fatal(err)
	}

	log.Println(u.String())
}
```

## 预签名上传

`GenerateUpload` 返回上传所需的方法、URL，以及附带的表单字段或请求头。

```go
package main

import (
	"context"
	"crypto/sha256"
	"log"
	"os"
	"time"

	"github.com/ix64/s3-go/s3up"
)

func main() {
	buf, err := os.ReadFile("./file.bin")
	if err != nil {
		log.Fatal(err)
	}

	sum := sha256.Sum256(buf)

	result, err := client.GenerateUpload(context.Background(), &s3up.GenerateParams{
		RemotePath:  "upload/file.bin",
		ExpireIn:    time.Minute,
		Size:        int64(len(buf)),
		ContentType: "application/octet-stream",
		Sha256:      sum[:],
		Metadata: map[string]string{
			"source": "web",
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("method=%s url=%s", result.Method, result.URL.String())
}
```

当 `Method` 为：

- `POST`：使用 `FormData` 组装 multipart/form-data 请求
- `PUT`：使用 `Header` 设置请求头后直接上传文件内容

## 配置

`s3.ParseConfig` 读取 JSON 配置。

### 最小配置

```json
{
  "endpoint": "http://127.0.0.1:9000",
  "bucket": "e2e-test",
  "bucket_lookup": "path",
  "access_key": "minio",
  "secret_key": "E2ETestPassword"
}
```

### 顶层配置字段

```json
{
  "endpoint": "https://s3.example.com",
  "bucket": "my-bucket",
  "bucket_lookup": "path",
  "prefix": "app-prod",
  "access_key": "AKIA...",
  "secret_key": "SECRET...",
  "upload_generator_type": "s3",
  "upload_generator_config": {},
  "download_generator_type": "s3",
  "download_generator_config": {}
}
```

字段说明：

- `endpoint`：对象存储服务地址
- `bucket`：bucket 名称
- `bucket_lookup`：bucket 寻址方式，支持 `dns`、`path`、`cname`
- `prefix`：对象 key 前缀
- `access_key` / `secret_key`：访问凭证
- `upload_generator_type`：上传生成器类型，默认 `s3`
- `download_generator_type`：下载生成器类型，默认 `s3`

## 下载生成器

### S3 下载生成器

```json
{
  "download_generator_type": "s3",
  "download_generator_config": {
    "endpoint": "https://s3.example.com",
    "bucket": "my-bucket",
    "bucket_lookup": "path",
    "region": "us-east-1",
    "prefix": "app-prod",
    "public_read": false,
    "access_key": "AKIA...",
    "secret_key": "SECRET..."
  }
}
```

### 阿里云 CDN 下载生成器

```json
{
  "download_generator_type": "aliyun_cdn",
  "download_generator_config": {
    "endpoint": "https://cdn.example.com",
    "prefix": "app-prod",
    "auth_mode": "type-a",
    "auth_key": "your-cdn-key",
    "dynamic_expire": true
  }
}
```

支持的 `auth_mode`：

- `type-a`
- `type-b`
- `type-c`
- `type-f`

### 腾讯云 CDN 下载生成器

```json
{
  "download_generator_type": "tencent_cloud_cdn",
  "download_generator_config": {
    "endpoint": "https://cdn.example.com",
    "prefix": "app-prod",
    "auth_mode": "type-a",
    "auth_key": "your-cdn-key",
    "dynamic_expire": true
  }
}
```

支持的 `auth_mode`：

- `type-a`
- `type-b`
- `type-c`
- `type-d`

## 上传生成器

### S3 上传生成器

```json
{
  "upload_generator_type": "s3",
  "upload_generator_config": {
    "endpoint": "https://s3.example.com",
    "bucket": "my-bucket",
    "bucket_lookup": "path",
    "region": "us-east-1",
    "prefix": "app-prod",
    "disable_post": false
  }
}
```

说明：

- 默认优先生成 Pre-signed POST
- 设置 `disable_post=true` 时回退到 Pre-signed PUT
- 某些 S3 兼容厂商不支持校验或 POST，可通过配置关闭对应能力

## 自定义生成器

你可以直接替换默认生成器：

```go
client.SetDownloadGenerator(customDownloadGenerator)
client.SetUploadGenerator(customUploadGenerator)
```

接口定义见：

- [s3down/interface.go](/d:/dev/s3-go/s3down/interface.go)
- [s3up/interface.go](/d:/dev/s3-go/s3up/interface.go)

## 测试

### 单元测试

```bash
go test ./...
```

### MinIO E2E

仓库内置了 MinIO 的 E2E 环境：

```bash
docker compose -f e2e/minio/docker-compose.yaml up -d
go test -v ./s3
docker compose -f e2e/minio/docker-compose.yaml down
```

或者直接使用 `Taskfile`：

```bash
task test-e2e:minio
```

E2E 配置文件默认使用 [e2e/minio/config.json](/d:/dev/s3-go/e2e/minio/config.json)。

如果要跑其他环境的 E2E，设置 `E2E_S3_CONFIG` 指向对应 JSON 文件即可。

## 已知限制

- 当前上传生成器仅支持 S3 兼容实现
- `bucket_lookup=cname` 目前不支持 S3 上传生成器
- E2E 测试依赖外部对象存储或 MinIO 环境

## License

[LICENSE](/d:/dev/s3-go/LICENSE)
