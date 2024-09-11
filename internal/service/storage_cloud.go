package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	transport "github.com/aws/smithy-go/endpoints"

	"gocloud.dev/blob"
	"gocloud.dev/blob/s3blob"
)

// AWS Crud
type Resolver struct {
	URL *url.URL
}

func (r *Resolver) ResolveEndpoint(_ context.Context, params s3.EndpointParameters) (transport.Endpoint, error) {
	u := *r.URL
	u.Path += "/" + *params.Bucket
	return transport.Endpoint{URI: u}, nil
}

//

type CloudStorage struct {
	BucketRef   *blob.Bucket
	BucketName  string
	Prefix      string
	StorageName string
}

const (
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

var (
	stringEmpty = func(key string) bool {
		return key == ""
	}
	initBucketOnce sync.Once
)

func (s *CloudStorage) GetPrefix() string {
	return s.Prefix
}

// getCloudLocation appends prefix to path
func (s *CloudStorage) getCloudLocation(fileName string) string {
	if s.Prefix == "<nil>" {
		s.Prefix = ""
	}
	// Skip if prefix is already in Path.
	if len(s.Prefix) > 0 && strings.Contains(fileName, s.Prefix) {
		return fileName
	}
	if fileName[0] != '/' && s.Prefix != "" {
		return path.Join(s.Prefix, "/", fileName)
	}
	return path.Join(s.Prefix, fileName)
}

// ReadFile read file from Cloud Provider and return byte array
func (s *CloudStorage) ReadFile(filename string) ([]byte, error) {
	if s.BucketRef == nil {
		return nil, errors.New("unable to find valid bucket to read file from")
	}
	ctx := context.Background()
	return s.BucketRef.ReadAll(ctx, s.getCloudLocation(filename))
}

// WriteFile persists data to Cloud Provider Storage returning error if operation failed
func (s *CloudStorage) WriteFile(filename string, data []byte) error {
	if s.BucketRef == nil {
		return errors.New("unable to get valid bucket ")
	}
	return s.BucketRef.WriteAll(context.Background(), s.getCloudLocation(filename), data, nil)
}

func (s *CloudStorage) Name() string {
	return s.StorageName
}

func (s *CloudStorage) FindAllFiles(folder string, fullPath bool) ([]string, error) {
	if s.BucketRef == nil {
		return nil, errors.New("unable to find valid bucket to list files from")
	}
	folderName := s.getCloudLocation(folder)

	var fileList []string
	opts := blob.ListOptions{}
	if s.Prefix != "" {
		opts.Prefix = folderName
	}

	iterator := s.BucketRef.List(&opts)
	for {
		obj, err := iterator.Next(context.Background())
		if err != nil {
			break
		}
		if fullPath {
			if strings.Contains(obj.Key, folderName) {
				fileList = append(fileList, obj.Key)
			} else {
				slog.Debug("key does not match folder path", "key", obj.Key)
			}
		} else {
			fileList = append(fileList, filepath.Base(obj.Key))
		}
	}

	return fileList, nil
}

func NewCloudStorage(c context.Context) (Storage, error) {
	var (
		err       error
		bucketObj *blob.Bucket
		errorMsg  string
	)

	contextVal := c.Value(StorageContext)
	if contextVal == nil {
		return nil, errors.New("cannot configure GCP storage, context missing")
	}
	appData, ok := contextVal.(map[string]string)
	if !ok {
		return nil, errors.New("cannot convert appData to string map")
	}

	// Pattern specifically for Self hosted S3 compatible instances Minio / Ceph
	if boolStrCheck(getMapValue(Custom, "false", stringEmpty, appData)) {
		creds := credentials.NewStaticCredentialsProvider(
			getMapValue(AccessId, os.Getenv("AWS_ACCESS_KEY"), stringEmpty, appData),
			getMapValue(SecretKey, os.Getenv("AWS_SECRET_KEY"), stringEmpty, appData), "")
		host := getMapValue(Endpoint, "http://localhost:9000", stringEmpty, appData)
		cloudCfg := &aws.Config{
			Credentials:  creds,
			Region:       getMapValue(Region, "us-east-1", stringEmpty, appData),
			BaseEndpoint: &host,
		}
		session := s3.NewFromConfig(*cloudCfg,
			func(o *s3.Options) {
				o.UsePathStyle = true //  <---- here
			},
			func(o *s3.Options) {
				endpointURL, _ := url.Parse(host) // or where ever you ran minio
				s3.WithEndpointResolverV2(&Resolver{URL: endpointURL})
			},
		)
		if session == nil {
			errorMsg = "No valid session could be created"
		}
		bucketObj, err = s3blob.OpenBucketV2(context.Background(), session, appData["bucket_name"], nil)
		if err != nil {
			errorMsg = err.Error()
		}
		if err == nil && boolStrCheck(getMapValue(InitBucket, "false", stringEmpty, appData)) {
			slog.Info("attempting to bootstrap bucket", slog.Any("bucket", appData[BucketName]))
			// Attempts to initiate bucket
			createBucket := func() {
				m := s3.CreateBucketInput{
					Bucket: aws.String(appData[BucketName]),
				}
				// attempt to create bucket
				_, err := session.CreateBucket(context.Background(), &m)
				if err != nil {
					slog.Warn("bucket already exists or cannot be created", "bucket", *m.Bucket)
				} else {
					slog.Info("bucket has been created", "bucket", *m.Bucket)
				}
			}

			if os.Getenv("TESTING") != "1" {
				initBucketOnce.Do(func() {
					createBucket()
				})
			} else {
				createBucket()
			}

		}

	} else {
		cloudURL := fmt.Sprintf("%s://%s", appData["cloud_type"], appData["bucket_name"])
		bucketObj, err = blob.OpenBucket(c, cloudURL)
		errorMsg = fmt.Sprintf("failed to open bucket %s", cloudURL)
	}

	if err != nil {
		log.Fatalf("unable to connect to cloud provider, err: %v, message: %s", err, errorMsg)
	}

	entity := &CloudStorage{
		BucketName: appData[BucketName],
		BucketRef:  bucketObj,
	}

	if val, ok := appData[Prefix]; ok {
		entity.Prefix = val
	}

	return entity, nil
}

// boolStrCheck does a more intelligent bool check as yaml values are converted to "1" or "true" depending
// on how the user configures quotes the value.
func boolStrCheck(val string) bool {
	return strings.ToLower(val) == "true" || val == "1"
}

// getMapValue a generic utility that will get a value from a map and return a default if key does not exist
func getMapValue[T comparable](key, defaultValue T, emptyTest func(key T) bool, data map[T]T) T {
	val, ok := data[key]
	if ok && !emptyTest(val) {
		return val
	}
	return defaultValue
}
