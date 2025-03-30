package s3

import (
	"context"
	"io"
	"mime"
	"net/url"
	"path"

	"github.com/minio/minio-go/v7"

	"github.com/ix64/s3-go/s3down"
	"github.com/ix64/s3-go/s3up"
)

// Upload 将 io.Reader 的内容上传到远程的文件
func (c *Client) Upload(ctx context.Context, remotePath string, file io.Reader, size int64, mime string) error {
	_, err := c.c.PutObject(ctx, c.cfg.Bucket, c.composeObjectName(remotePath), file, size, minio.PutObjectOptions{ContentType: mime})
	return err
}

// UploadFile 将本地文件上传到远程
func (c *Client) UploadFile(ctx context.Context, remotePath string, localPath string) error {
	_, err := c.c.FPutObject(ctx,
		c.cfg.Bucket,
		c.composeObjectName(remotePath),
		localPath,
		minio.PutObjectOptions{
			ContentType: mime.TypeByExtension(path.Ext(remotePath)),
		},
	)
	return err
}

// Download 获取文件内容，返回 io.ReadCloser
func (c *Client) Download(ctx context.Context, remotePath string) (io.ReadCloser, error) {
	return c.c.GetObject(ctx,
		c.cfg.Bucket,
		c.composeObjectName(remotePath),
		minio.GetObjectOptions{},
	)
}

// DownloadFile 下载文件到指定的本地路径
func (c *Client) DownloadFile(ctx context.Context, remotePath string, localPath string) error {
	return c.c.FGetObject(ctx,
		c.cfg.Bucket,
		c.composeObjectName(remotePath),
		localPath,
		minio.GetObjectOptions{},
	)
}

// Stat 获取文件信息
func (c *Client) Stat(ctx context.Context, remotePath string) (minio.ObjectInfo, error) {
	return c.c.StatObject(ctx, c.cfg.Bucket, c.composeObjectName(remotePath), minio.StatObjectOptions{
		Checksum: true,
	})
}

// Delete 删除文件
func (c *Client) Delete(ctx context.Context, remotePath string) error {
	return c.c.RemoveObject(ctx, c.cfg.Bucket, c.composeObjectName(remotePath), minio.RemoveObjectOptions{})
}

// Copy 远程复制文件
func (c *Client) Copy(ctx context.Context, oldPath string, newPath string) error {
	srcOpts := minio.CopySrcOptions{Bucket: c.cfg.Bucket, Object: c.composeObjectName(oldPath)}
	dstOpts := minio.CopyDestOptions{Bucket: c.cfg.Bucket, Object: c.composeObjectName(newPath)}

	_, err := c.c.CopyObject(ctx, dstOpts, srcOpts)
	return err
}

// Move 远程移动文件（复制后删除）
func (c *Client) Move(ctx context.Context, oldPath, newPath string) error {
	if err := c.Copy(ctx, oldPath, newPath); err != nil {
		return err
	}
	return c.Delete(ctx, oldPath)
}

// GenerateDownload 前端直连下载 预签名生成下载链接
func (c *Client) GenerateDownload(ctx context.Context, params *s3down.GenerateParams) (*url.URL, error) {
	return c.download.GenerateDownload(ctx, params)
}

// GenerateUpload 前端直连上传 预签名生成上传链接
func (c *Client) GenerateUpload(ctx context.Context, param *s3up.GenerateParams) (*s3up.GenerateResult, error) {
	return c.upload.GenerateUpload(ctx, param)
}
