package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	log "github.com/sirupsen/logrus"
	"gocloud.dev/blob"
	"gocloud.dev/blob/s3blob"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
)

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

// getCloudLocation appends prefix to path
func (s *CloudStorage) getCloudLocation(fileName string) string {
	if s.Prefix == "<nil>" {
		s.Prefix = ""
	}
	//Skip if prefix is already in Path.
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
				log.Debugf("%s does not match folder path", obj.Key)
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

	//Pattern specifically for Self hosted S3 compatible instances Minio / Ceph
	if boolStrCheck(getMapValue(Custom, "false", stringEmpty, appData)) {
		var sess *session.Session
		creds := credentials.NewStaticCredentials(
			getMapValue(AccessId, os.Getenv("AWS_ACCESS_KEY"), stringEmpty, appData),
			getMapValue(SecretKey, os.Getenv("AWS_SECRET_KEY"), stringEmpty, appData), "")
		sess, err = session.NewSession(&aws.Config{
			Credentials:      creds,
			Endpoint:         aws.String(getMapValue(Endpoint, "http://localhost:9000", stringEmpty, appData)),
			DisableSSL:       aws.Bool(getMapValue(SSLEnabled, "false", stringEmpty, appData) != "true"),
			S3ForcePathStyle: aws.Bool(true),
			Region:           aws.String(getMapValue(Region, "us-east-1", stringEmpty, appData)),
		})
		if err != nil {
			errorMsg = err.Error()
		}
		bucketObj, err = s3blob.OpenBucket(context.Background(), sess, appData["bucket_name"], nil)
		if err != nil {
			errorMsg = err.Error()
		}
		if err == nil && boolStrCheck(getMapValue(InitBucket, "false", stringEmpty, appData)) {
			//Attempts to initiate bucket
			initBucketOnce.Do(func() {
				client := s3.New(sess)
				m := s3.CreateBucketInput{
					Bucket: aws.String(appData[BucketName]),
				}
				//attempt to create bucket
				_, err := client.CreateBucket(&m)
				if err != nil {
					log.Warnf("%s bucket already exists or cannot be created", *m.Bucket)
				} else {
					log.Infof("bucket %s has been created", *m.Bucket)
				}
			})

		}

	} else {
		var cloudURL = fmt.Sprintf("%s://%s", appData["cloud_type"], appData["bucket_name"])
		bucketObj, err = blob.OpenBucket(c, cloudURL)
		errorMsg = fmt.Sprintf("failed to open bucket %s", cloudURL)
	}

	if err != nil {
		log.WithError(err).WithField("Msg", errorMsg).Fatal("unable to connect to cloud provider")
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
