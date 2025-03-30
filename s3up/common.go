package s3up

import (
	"path"
	"strings"
)

func composeObjectName(prefix string, remotePath string) string {
	// s3 object name can not start with "/"
	return strings.TrimPrefix(path.Join(prefix, remotePath), "/")
}
