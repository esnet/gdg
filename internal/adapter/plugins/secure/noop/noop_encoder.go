package noop

import (
	"github.com/esnet/gdg/internal/domain"
)

type NoOpEncoder struct{}

func (n NoOpEncoder) EncodeValue(b string) (string, error) {
	return b, nil
}

func (n NoOpEncoder) DecodeValue(b string) (string, error) {
	return b, nil
}

func (n NoOpEncoder) Encode(resourceType domain.ResourceType, b []byte) ([]byte, error) {
	return b, nil
}

func (n NoOpEncoder) Decode(resourceType domain.ResourceType, b []byte) ([]byte, error) {
	return b, nil
}
