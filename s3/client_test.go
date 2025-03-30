package s3_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/ix64/s3-go/s3down"
	"github.com/ix64/s3-go/s3up"
)

func TestClient_UploadDownloadFile(t *testing.T) {
	c, tempDir, err := initE2EClient(t)
	if err != nil {
		t.Errorf("failed to init e2e client: %v", err)
	}

	origPath := filepath.Join(tempDir, localPathRandom)
	downloadPath := filepath.Join(tempDir, "ops-file_download.bin")
	remotePath := "__e2e_test__/ops-file/uploaded.bin"

	t.Run("UploadFile", func(t *testing.T) {
		err := c.UploadFile(context.Background(), remotePath, origPath)
		assert.NoError(t, err)
	})

	t.Run("DownloadFile", func(t *testing.T) {
		err = c.DownloadFile(context.Background(), remotePath, downloadPath)
		assert.NoError(t, err)
	})

	t.Run("DeleteFile", func(t *testing.T) {
		err = c.Delete(context.Background(), remotePath)
		assert.NoError(t, err)
	})

	t.Run("FileContentSame", func(t *testing.T) {
		assertFileContentSame(t, origPath, downloadPath)
	})

}

func TestClient_ObjectOperation(t *testing.T) {
	c, tempDir, err := initE2EClient(t)
	if err != nil {
		t.Errorf("failed to init e2e client: %v", err)
	}

	origPath := filepath.Join(tempDir, localPathRandom)
	origStat, err := os.Stat(origPath)
	if err != nil {
		t.Fatal(err)
	}
	downloadPath := filepath.Join(tempDir, "ops-object_download.bin")
	remotePath := "__e2e_test__/ops-object/uploaded.bin"

	t.Run("Upload", func(t *testing.T) {
		f, err := os.Open(origPath)
		assert.NoError(t, err)
		defer f.Close()

		err = c.Upload(context.Background(), remotePath, f, origStat.Size(), "application/octet-stream")
		assert.NoError(t, err)
	})

	t.Run("Stat", func(t *testing.T) {
		stat, err := c.Stat(context.Background(), remotePath)
		assert.NoError(t, err)
		assert.Equal(t, origStat.Size(), stat.Size)
	})

	t.Run("Download", func(t *testing.T) {
		f, err := os.Create(downloadPath)
		assert.NoError(t, err)
		defer f.Close()

		r, err := c.Download(context.Background(), remotePath)
		assert.NoError(t, err)

		_, err = io.Copy(f, r)
		assert.NoError(t, err)
	})

	t.Run("Delete", func(t *testing.T) {
		err = c.Delete(context.Background(), remotePath)
		assert.NoError(t, err)
	})

	t.Run("FileContentSame", func(t *testing.T) {
		assertFileContentSame(t, origPath, downloadPath)
	})

}

func TestClient_GenerateDownload(t *testing.T) {
	c, tempDir, err := initE2EClient(t)
	if err != nil {
		t.Errorf("failed to init e2e client: %v", err)
	}

	origPath := filepath.Join(tempDir, "/random.bin")
	downloadPath := filepath.Join(tempDir, "generate-download_download.bin")
	remotePath := "__e2e_test__/generate-download/uploaded.bin"

	t.Run("GenerateDownload", func(t *testing.T) {
		err := c.UploadFile(context.Background(), remotePath, origPath)
		assert.NoError(t, err)

		// cleanup remote file if uploaded, whether success or not
		t.Cleanup(func() {
			err := c.Delete(context.Background(), remotePath)
			assert.NoError(t, err)
		})

		url, err := c.GenerateDownload(context.Background(), &s3down.GenerateParams{
			RemotePath:         remotePath,
			ExpireIn:           time.Minute,
			ContentType:        "application/octet-stream",
			AttachmentFilename: "random.bin",
		})
		assert.NoError(t, err)

		req, err := http.NewRequest(http.MethodGet, url.String(), nil)
		assert.NoError(t, err)

		t.Logf("prepared url: %s", url.String())

		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		f, err := os.Create(downloadPath)
		assert.NoError(t, err)
		defer f.Close()

		_, err = io.Copy(f, resp.Body)
		assert.NoError(t, err)
	})

	t.Run("FileContentSame", func(t *testing.T) {
		assertFileContentSame(t, origPath, downloadPath)
	})

}

func TestClient_GenerateUpload(t *testing.T) {
	c, tempDir, err := initE2EClient(t)
	if err != nil {
		t.Errorf("failed to init e2e client: %v", err)
	}

	origPath := filepath.Join(tempDir, localPathRandom)
	origStat, err := os.Stat(origPath)
	assert.NoError(t, err)

	origBuf, err := os.ReadFile(origPath)
	assert.NoError(t, err)

	origSha256 := sha256.Sum256(origBuf)

	downloadPath := filepath.Join(tempDir, "generate-upload_download.bin")
	remotePath := "/__unit_test/generate-upload/random.bin"

	t.Run("GenerateUpload", func(t *testing.T) {

		info, err := c.GenerateUpload(t.Context(), &s3up.GenerateParams{
			RemotePath:  remotePath,
			ExpireIn:    time.Minute,
			Size:        origStat.Size(),
			ContentType: "application/octet-stream",
			Sha256:      origSha256[:],
			Metadata: map[string]string{
				"foo": "bar",
			},
		})
		assert.NoError(t, err)

		req, err := generateUserUploadRequest(info, origPath)
		assert.NoError(t, err)

		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		defer resp.Body.Close()

		assert.True(t, resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices)

		// cleanup remote file if uploaded, whether success or not
		t.Cleanup(func() {
			err := c.Delete(context.Background(), remotePath)
			assert.NoError(t, err)
		})

		err = c.DownloadFile(context.Background(), remotePath, downloadPath)
		assert.NoError(t, err)
	})

	t.Run("FileContentSame", func(t *testing.T) {
		assertFileContentSame(t, origPath, downloadPath)
	})
}

func assertFileContentSame(t *testing.T, path1, path2 string) {
	f1, err := os.ReadFile(path1)
	assert.NoError(t, err)

	f2, err := os.ReadFile(path2)
	assert.NoError(t, err)

	sum1 := sha256.Sum256(f1)
	sum2 := sha256.Sum256(f2)

	assert.Equal(t, sum1[:], sum2[:])
}

func generateUserUploadRequest(info *s3up.GenerateResult, localPath string) (*http.Request, error) {
	buf := bytes.NewBuffer(nil)

	w := multipart.NewWriter(buf)
	defer w.Close()

	for k, v := range info.FormData {
		if err := w.WriteField(k, v); err != nil {
			return nil, err
		}
	}

	fw, err := w.CreateFormFile("file", filepath.Base(localPath))
	if err != nil {
		return nil, err
	}

	f, err := os.Open(localPath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	if _, err := io.Copy(fw, f); err != nil {
		return nil, err
	}

	if err := w.Close(); err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, info.URL.String(), buf)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", w.FormDataContentType())

	return req, nil
}
