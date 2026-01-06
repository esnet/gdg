package contract

import "github.com/esnet/gdg/pkg/config/domain"

const (
	EncodeOperation = "Encode"
	DecodeOperation = "Decode"
)

type CipherEncoder interface {
	Encode(resourceType domain.ResourceType, b []byte) ([]byte, error)
	Decode(resourceType domain.ResourceType, b []byte) ([]byte, error)
	EncodeValue(b string) (string, error)
	DecodeValue(b string) (string, error)
}
