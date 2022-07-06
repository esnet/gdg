package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"syscall"

	"github.com/graymeta/stow"
	log "github.com/sirupsen/logrus"
)

type CloudStorage struct {
	Location    stow.Location
	BucketName  string
	BucketRef   stow.Container
	Prefix      string
	StorageName string
}

const (
	CloudType  = "cloud_type"
	BucketName = "bucket_name"
	Prefix     = "prefix"
)

//getCloudLocation appends prefix to path
func (s *CloudStorage) getCloudLocation(fileName string) string {
	if s.Prefix == "<nil>" {
		s.Prefix = ""
	}
	if fileName[len(fileName)-2] != '/' && s.Prefix != "" {
		return path.Join(s.Prefix, "/", fileName)
	}
	return path.Join(s.Prefix, fileName)
}

//ReadFile read file from Cloud Provider and return byte array
func (s *CloudStorage) ReadFile(filename string) ([]byte, error) {
	item, err := s.BucketRef.Item(s.getCloudLocation(filename))
	if err != nil {
		return nil, errors.New("file not found on Cloud")
	}
	r, err := item.Open()
	defer func() {
		err = r.Close()
		if err != nil {
			log.Error("Failed to close Cloud file")
		}
	}()
	if err != nil {
		return nil, err
	}
	return io.ReadAll(r)
}

//WriteFile persists data to Cloud Provider Storage returning error if operation failed
func (s *CloudStorage) WriteFile(filename string, data []byte, mode fs.FileMode) error {
	reader := bytes.NewReader(data)
	size := int64(len(data))
	item, err := s.BucketRef.Put(s.getCloudLocation(filename), reader, size, nil)
	if err != nil {
		log.WithError(err).Errorf("failed to write %s to Cloud at location: %s", filename, item.URL())
		return err
	}
	return nil
}

func (s CloudStorage) Name() string {
	return s.StorageName
}

func (s CloudStorage) FindAllFiles(folder string, fullPath bool) ([]string, error) {
	folderName := s.getCloudLocation(folder)
	var result []string
	err := stow.Walk(s.BucketRef, folderName, 100, func(c stow.Item, err error) error {
		if err != nil {
			return err
		}
		if c != nil {
			if fullPath {
				result = append(result, c.Name())
			} else {
				result = append(result, filepath.Base(c.Name()))
			}
			return nil
		} else {
			return errors.New("could not append file")
		}
	})

	return result, err
}

func NewCloudStorage(c context.Context) Storage {
	var (
		item     stow.Container
		location stow.Location
		err      error
		data     []byte
	)
	contextVal := c.Value(StorageContext)
	if contextVal == nil {
		log.Fatal("Cannot configure GCP storage, context missing")
	}
	appData, ok := contextVal.(map[string]interface{})
	if !ok {
		log.Fatal("Cannot convert appData to string map")
	}
	config := stow.ConfigMap{}
	for key, value := range appData {
		stringVal := fmt.Sprintf("%v", value)
		if stringVal == "<nil>" {
			stringVal = ""
		}
		config[key] = stringVal
	}

	jsonKey, ok := config["json"]
	if ok && !isJSON(jsonKey) {
		data, err = ioutil.ReadFile(jsonKey)
		if err != nil {
			log.WithError(err).Errorf("Failed to read service key file")
		}
		config["json"] = string(data)
	}

	location, err = stow.Dial(config["cloud_type"], config)
	if err != nil {
		log.WithError(err).Fatal("Unable to connect to Cloud")
	}
	entity := &CloudStorage{
		Location:   location,
		BucketName: config[BucketName],
	}
	entity.Prefix = config[Prefix]

	entity.BucketRef, err = location.Container(entity.BucketName)
	if err != nil {
		log.WithError(err).Fatalf("Bucket %s is either not found or not accessible", item.Name())
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		err = location.Close()
		if err != nil {
			log.Error("Failed to close location")
		}
		os.Exit(0)
	}()

	return entity
}

//isJSON Utility function to test if string is a JSON object
func isJSON(s string) bool {
	var js map[string]interface{}
	return json.Unmarshal([]byte(s), &js) == nil

}
