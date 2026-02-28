package grafana

import (
	"github.com/esnet/gdg/internal/domain"
	"github.com/grafana/grafana-openapi-client-go/models"
)

func WithNestedToCreateLibraryElement(entry domain.WithNested[*models.LibraryElementDTO]) *models.CreateLibraryElementCommand {
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
