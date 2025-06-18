package types

import (
	"github.com/grafana/grafana-openapi-client-go/models"
)

func WithNestedToCreateLibraryElement(entry WithNested[*models.LibraryElementDTO]) *models.CreateLibraryElementCommand {
	data := *entry.Entity
	obj := &models.CreateLibraryElementCommand{
		FolderUID: data.FolderUID,
		Kind:      data.Kind,
		Model:     data.Model,
		Name:      data.Name,
		UID:       data.UID,
	}
	return obj
}
