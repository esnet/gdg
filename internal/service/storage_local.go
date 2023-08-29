package service

import (
	"context"
	"errors"
	log "github.com/sirupsen/logrus"
	"gocloud.dev/blob"
	"gocloud.dev/blob/fileblob"
	_ "gocloud.dev/blob/fileblob"
	"os"
	"path/filepath"
)

// LocalStorage default storage engine
type LocalStorage struct {
}

// ReadFile returns a byte array of file content
func (s *LocalStorage) ReadFile(filename string) ([]byte, error) {
	mb, err := s.getBucket(filepath.Dir(filename))
	if err != nil {
		return nil, err
	}
	f := filepath.Base(filename)
	data, err := mb.ReadAll(context.Background(), f)
	if err != nil || len(data) == 0 {
		return nil, errors.New("unable to read file")
	}
	return data, nil
}

func (s *LocalStorage) getBucket(baseFolder string) (*blob.Bucket, error) {
	if _, err := os.Stat(baseFolder); err != nil {
		_ = os.Mkdir(baseFolder, 0750)
	}
	return fileblob.OpenBucket(baseFolder, nil)
}

// WriteFile writes file to disk and returns an error if operation failed
func (s *LocalStorage) WriteFile(filename string, data []byte) error {
	mb, err := s.getBucket(filepath.Dir(filename))
	if err != nil {
		return err
	}
	f := filepath.Base(filename)
	err = mb.WriteAll(context.Background(), f, data, nil)
	if err == nil {
		//Remove attribute file being generated by local storage
		attrFile := filename + ".attrs"
		log.Debugf("Removing file %s", attrFile)
		defer os.Remove(attrFile)

	}
	return err
}

func (LocalStorage) Name() string {
	return "LocalStorage"
}

func (s *LocalStorage) FindAllFiles(folder string, fullPath bool) ([]string, error) {
	mb, err := s.getBucket(folder)
	if err != nil {
		return nil, err
	}

	var fileList []string
	iterator := mb.List(nil)
	for {
		obj, err := iterator.Next(context.Background())
		if err != nil {
			break
		}
		if fullPath {
			fileList = append(fileList, filepath.Join(folder, obj.Key))
		} else {
			fileList = append(fileList, filepath.Base(obj.Key))
		}
	}

	return fileList, nil
}

func NewLocalStorage(ctx context.Context) Storage {
	return &LocalStorage{}
}
