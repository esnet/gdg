package api

import (
	"context"
	"errors"
	log "github.com/sirupsen/logrus"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
)

//LocalStorage default storage engine
type LocalStorage struct {
}

//ReadFile returns a byte array of file content
func (s *LocalStorage) ReadFile(filename string) ([]byte, error) {
	return ioutil.ReadFile(filename)
}

//WriteFile writes file to disk and returns an error if operation failed
func (s *LocalStorage) WriteFile(filename string, data []byte, mode fs.FileMode) error {
	return ioutil.WriteFile(filename, data, mode)
}

func (LocalStorage) Name() string {
	return "LocalStorage"
}

func (s *LocalStorage) FindAllFiles(folder string) ([]string, error) {
	if _, err := os.Stat(folder); os.IsNotExist(err) {
		log.Warn("Output folder was not found")
		return []string{}, errors.New("unable to find requested folder")
	}
	var fileList []string
	err := filepath.Walk(folder, func(path string, f os.FileInfo, err error) error {
		if !f.IsDir() {
			fileList = append(fileList, path)
		}
		return nil
	})
	if err != nil {
		panic(err)
	}

	return fileList, nil
}

//ReadDir returns the basename of all files
func (s *LocalStorage) ReadDir(dir string) ([]string, error) {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var result []string
	for _, file := range files {
		result = append(result, file.Name())
	}

	return result, nil
}

func NewLocalStorage(ctx context.Context) Storage {
	return &LocalStorage{}
}
