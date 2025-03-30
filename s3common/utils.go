package s3common

import (
	"net/url"
	"strconv"
	"strings"
)

func ComposeContentDisposition(filename string) string {
	return strings.Join([]string{
		"attachment",

		// RFC6266: should be ISO-8859-1 only, but most browsers support UTF-8
		"filename=" + strconv.Quote(filename),

		// RFC6266: UTF-8 support
		"filename*=UTF-8''" + url.QueryEscape(filename),
	}, "; ")
}
