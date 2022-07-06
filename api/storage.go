package api

import (
	_ "github.com/graymeta/stow/azure"
	_ "github.com/graymeta/stow/google"
	_ "github.com/graymeta/stow/s3"
	_ "github.com/graymeta/stow/sftp"
	"io/fs"
)

type ContextStorage string

const StorageContext = ContextStorage("storage")

//TODO: pull all the cloud based interaction into a Plugin System
type Storage interface {
	WriteFile(filename string, data []byte, mode fs.FileMode) error // WriteFile returns error or writes byte array to destination
	ReadFile(filename string) ([]byte, error)                       // ReadFile returns byte array or error with data from file
	FindAllFiles(folder string, fullPath bool) ([]string, error)    // FindAllFiles recursively list all files for a given path
	Name() string                                                   // Name of storage engine
}
