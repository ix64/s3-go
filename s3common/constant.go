package s3common

type BucketLookupType string

const (

	// BucketLookupDNS prepend bucket name to endpoint host
	BucketLookupDNS = "dns"

	// BucketLookupPath prepend bucket name to endpoint path
	BucketLookupPath = "path"

	// BucketLookupCNAME does not change endpoint, custom domain has been mapped to bucket
	BucketLookupCNAME = "cname"
)
