package storage

type ContextStorage string

type Type string

func (t Type) String() string {
	return string(t)
}

const (
	Context = ContextStorage("storage")
	// Cloud Specific const
	CloudType             = "cloud_type"
	BucketName            = "bucket_name"
	Prefix                = "prefix"
	Custom                = "custom"
	AccessId              = "access_id"
	SecretKey             = "secret_key"
	Endpoint              = "endpoint"
	Region                = "region"
	InitBucket            = "init_bucket"
	LocalStorageType Type = "LocalStorage"
)
