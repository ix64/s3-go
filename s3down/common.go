package s3down

import (
	"net/url"
	"path"
	"strconv"
	"strings"
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

func composeContentDisposition(filename string) string {
	return strings.Join([]string{
		"attachment",

		// RFC6266: should be ISO-8859-1 only, but most browsers support UTF-8
		"filename=" + strconv.Quote(filename),

		// RFC6266: UTF-8 support
		"filename*=UTF-8''" + url.QueryEscape(filename),
	}, "; ")
}

var TimezoneCST = time.FixedZone("CST", 8*60*60)
