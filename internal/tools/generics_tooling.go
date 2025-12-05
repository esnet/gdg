package tools

import (
	"encoding/json"
	"log"
	"log/slog"
	"os"
	"sync"
)

// DeepCopy creates a deep copy of the given value using JSON serialization, returning a pointer and an error.
func DeepCopy[T any](value T) (*T, error) {
	origJSON, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}

	clone := new(T)
	if err = json.Unmarshal(origJSON, &clone); err != nil {
		return nil, err
	}

	return clone, nil
}

var syncMap = new(sync.Map)

// CreateDestinationPath Handle osMkdir Errors
func CreateDestinationPath(folderName string, clearOutput bool, v string) {
	if clearOutput {
		// ensure the folder is only removed once.  This prevents valid data from being removed.
		_, ok := syncMap.Load(folderName)
		if !ok {
			syncMap.Store(folderName, true)
			clearBackup := os.RemoveAll(folderName)
			if clearBackup != nil {
				slog.Warn("Unable to remove previous backup at location", "location", v)
			}
		}
	}

	err := os.MkdirAll(v, 0o750)
	if err != nil {
		log.Fatalf("unable to create path %s, err: %s", v, err.Error())
	}
}
