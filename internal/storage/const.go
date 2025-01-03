package storage

type ContextStorage string

const (
	Context = ContextStorage("storage")
	// Cloud Specific const
	CloudType  = "cloud_type"
	BucketName = "bucket_name"
	Prefix     = "prefix"
	Kind       = "kind"
	Custom     = "custom"
	AccessId   = "access_id"
	SecretKey  = "secret_key"
	Endpoint   = "endpoint"
	Region     = "region"
	SSLEnabled = "ssl_enabled"
	InitBucket = "init_bucket"
)
