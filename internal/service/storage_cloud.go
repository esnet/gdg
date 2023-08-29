package service

import (
	"context"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"gocloud.dev/blob"
	"path"
	"path/filepath"
	"strings"
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
			fileList = append(fileList, obj.Key)
		} else {
			fileList = append(fileList, filepath.Base(obj.Key))
		}
	}

	return fileList, nil
}

func NewCloudStorage(c context.Context) (Storage, error) {
	var (
		err error
	)

	contextVal := c.Value(StorageContext)
	if contextVal == nil {
		return nil, errors.New("cannot configure GCP storage, context missing")
	}
	appData, ok := contextVal.(map[string]string)
	if !ok {
		return nil, errors.New("cannot convert appData to string map")
	}

	var cloudURL = fmt.Sprintf("%s://%s", appData["cloud_type"], appData["bucket_name"])

	bucketObj, err := blob.OpenBucket(c, cloudURL)
	if err != nil {
		log.Panicf("failed to open bucket %s", cloudURL)
	}

	config := map[string]string{}

	for key, value := range appData {
		stringVal := fmt.Sprintf("%v", value)
		if stringVal == "<nil>" {
			stringVal = ""
		}
		config[key] = stringVal
	}

	entity := &CloudStorage{
		BucketName: config[BucketName],
		BucketRef:  bucketObj,
	}
	if val, ok := config["prefix"]; ok {
		entity.Prefix = val
	}

	return entity, nil
}
