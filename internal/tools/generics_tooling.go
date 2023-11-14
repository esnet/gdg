package tools

import (
	"encoding/json"
	"log"
	"os"
)

func PtrOf[T any](value T) *T {
	return &value
}

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

// CreateDestinationPath Handle osMkdir Errors
func CreateDestinationPath(v string) {
	err := os.MkdirAll(v, 0750)
	if err != nil {
		log.Fatalf("unable to create path %s, err: %s", v, err.Error())
	}
}
