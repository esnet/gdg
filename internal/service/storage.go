package service

import (
	_ "gocloud.dev/blob/azureblob"
	_ "gocloud.dev/blob/gcsblob"
	_ "gocloud.dev/blob/s3blob"
	"io/fs"
)

type ContextStorage string

const StorageContext = ContextStorage("storage")

// TODO: pull all the cloud based interaction into a Plugin System
type Storage interface {
	WriteFile(filename string, data []byte, mode fs.FileMode) error // WriteFile returns error or writes byte array to destination
	ReadFile(filename string) ([]byte, error)                       // ReadFile returns byte array or error with data from file
	FindAllFiles(folder string, fullPath bool) ([]string, error)    // FindAllFiles recursively list all files for a given path
	Name() string                                                   // Name of storage engine
}
