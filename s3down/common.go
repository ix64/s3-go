package s3down

import (
	"net/url"
	"path"
	"time"
)

func composeObjectURL(base *url.URL, prefix, remotePath string) *url.URL {
	ret := *base // copy
	ret.Path = path.Join(
		ret.Path,
		prefix,
		path.Clean("/"+remotePath), // clean to avoid escape
	)
	return &ret
}

var TimezoneCST = time.FixedZone("CST", 8*60*60)
