package s3_test

import (
	"crypto/rand"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/tailscale/hujson"

	"github.com/ix64/s3-go/s3"
)

const (
	localPathRandom = "random.bin"
)

func initE2EClient(t *testing.T) (client *s3.Client, tempDir string, err error) {

	{ // load config & create client
		cfgPath, ok := os.LookupEnv("E2E_S3_CONFIG")
		if !ok || cfgPath == "" {
			t.Skipf("env E2E_S3_CONFIG not set")
		}

		cfgContent, err := os.ReadFile(cfgPath)
		if err != nil {
			return nil, "", fmt.Errorf("failed to read config file: %w", err)
		}

		cfgContent, err = hujson.Standardize(cfgContent)
		if err != nil {
			return nil, "", fmt.Errorf("failed to standardize config: %w", err)
		}

		cfg, err := s3.ParseConfig(cfgContent)
		if err != nil {
			return nil, "", fmt.Errorf("failed to parse config: %w", err)
		}

		client, err = s3.NewClient(cfg)
		if err != nil {
			return nil, "", fmt.Errorf("failed to create client: %w", err)
		}
	}

	{
		tempDir, err = os.MkdirTemp("", "")
		if err != nil {
			return nil, "", fmt.Errorf("failed to create temp dir: %w", err)
		}

		tempFile, err := os.Create(filepath.Join(tempDir, localPathRandom))
		if err != nil {
			return nil, "", fmt.Errorf("failed to create temp file: %w", err)
		}
		defer tempFile.Close()

		_, err = io.CopyN(tempFile, rand.Reader, 1*1024) // 1KB
		if err != nil {
			return nil, "", fmt.Errorf("failed to write to temp file: %w", err)
		}

		t.Cleanup(func() {
			_ = os.RemoveAll(tempDir)
		})

	}

	return client, tempDir, nil
}
