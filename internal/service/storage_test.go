package service

import (
	"context"
	"errors"
	"log"
	"path"
)

//TestStorage implements a very basic mock style storage that utilizes context and a hashmap to drive all the seed data.
// the purpose of this pattern is mainly ensuring that the expectations of the type of return from the various methods
// will result in a valid run.

// Aka ReadDir returns a list of file with only the basename, given that we should be able to find the appropriate file
// using app convention and established Prefix in order to drive the application.  If any CRUD operation files, then
// the test is consistent.
type TestStorage struct {
	m      map[string]interface{}
	Prefix string
}

func (s *TestStorage) getTestFileLocation(fileName string) string {
	if fileName[0] != '/' {
		return path.Join(s.Prefix, "/", fileName)
	}
	return path.Join(s.Prefix, fileName)
}

func (s *TestStorage) ReadFile(filename string) ([]byte, error) {
	val, ok := s.m[`filedata`]
	if ok {
		readCache := val.(map[string][]byte)
		return readCache[filename], nil
	}
	return nil, errors.New("file not found")

}

func (s *TestStorage) WriteFile(filename string, data []byte) error {
	val, ok := s.m[`filedata`]
	if ok {
		writeCache := val.(map[string][]byte)
		writeCache[filename] = data
	}

	return errors.New("Cannot save file")
}

func (s *TestStorage) FindAllFiles(folder string, fullPath bool) ([]string, error) {
	folderName := s.getTestFileLocation(folder)
	val, ok := s.m["fileList"]
	if ok {
		data := val.(map[string][]string)
		return data[folderName], nil
	}

	return nil, errors.New("No file list test data was found")
}

func (TestStorage) Name() string {
	return "Testing"
}

func NewTestStorage(c context.Context) Storage {
	contextValue := c.Value("seeddata").(map[string]interface{})
	if contextValue == nil {
		log.Fatal("Cannot configure Test storage, context missing")
	}
	prefix := c.Value("prefix")
	entity := &TestStorage{}
	entity.m = contextValue
	prefixValue, ok := prefix.(string)
	if ok {
		entity.Prefix = prefixValue
	}
	val, ok := entity.m["filedata"]
	if !ok || val == nil {
		entity.m["filedata"] = make(map[string][]byte)
	}
	val, ok = entity.m["fileList"]
	if !ok || val == nil {
		entity.m["fileList"] = make(map[string][]string)
	}
	return entity
}
