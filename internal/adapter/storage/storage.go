package storage

import (
	"context"
	"fmt"

	"github.com/esnet/gdg/internal/ports"
	_ "gocloud.dev/blob/azureblob"
	_ "gocloud.dev/blob/gcsblob"
	_ "gocloud.dev/blob/s3blob"
)

func NewStorageFromConfig(storageType string, appData map[string]string, encoder ports.CipherEncoder) (ports.Storage, error) {
	var (
		storageEngine ports.Storage
		err           error
	)

	ctx := context.Background()
	ctx = context.WithValue(ctx, Context, appData)
	switch storageType {
	case "cloud":
		{
			storageEngine, err = NewCloudStorage(ctx, encoder)
			if err != nil {
				return nil, fmt.Errorf("unable to configure CloudStorage Engine:	%w", err)
			}
		}
	default:
		storageEngine = NewLocalStorage(ctx)
	}
	return storageEngine, nil
}
