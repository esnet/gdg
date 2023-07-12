package tools

import "encoding/json"

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
