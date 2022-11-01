package api

import (
	"context"
	"errors"
	log "github.com/sirupsen/logrus"
	"io/fs"
	"os"
	"path/filepath"
)

//LocalStorage default storage engine
type LocalStorage struct {
}

//ReadFile returns a byte array of file content
func (s *LocalStorage) ReadFile(filename string) ([]byte, error) {
	return os.ReadFile(filename)
}

//ReadFileAbsolute returns the byte array of file content ignoring the prefix
func (s *LocalStorage) ReadFileAbsolute(filename string) ([]byte, error) {
	return s.ReadFile(filename)
}

//WriteFile writes file to disk and returns an error if operation failed
func (s *LocalStorage) WriteFile(filename string, data []byte, mode fs.FileMode) error {
	return os.WriteFile(filename, data, mode)
}

func (LocalStorage) Name() string {
	return "LocalStorage"
}

func (s *LocalStorage) FindAllFiles(folder string, fullPath bool) ([]string, error) {
	if _, err := os.Stat(folder); os.IsNotExist(err) {
		log.Warn("Output folder was not found")
		return []string{}, errors.New("unable to find requested folder")
	}
	var fileList []string
	err := filepath.Walk(folder, func(path string, f os.FileInfo, err error) error {
		if !f.IsDir() {
			if fullPath {
				fileList = append(fileList, path)
			} else {
				fileList = append(fileList, filepath.Base(path))
			}
		}
		return nil
	})
	if err != nil {
		panic(err)
	}

	return fileList, nil
}

func NewLocalStorage(ctx context.Context) Storage {
	return &LocalStorage{}
}
