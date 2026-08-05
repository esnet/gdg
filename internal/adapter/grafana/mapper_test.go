package grafana

import (
	"testing"

	"github.com/esnet/gdg/internal/domain"
	"github.com/grafana/grafana-openapi-client-go/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithNestedToCreateLibraryElement_AllFields(t *testing.T) {
	dto := &models.LibraryElementDTO{
		FolderUID: "folder-uid-123",
		Kind:      1,
		Model:     map[string]any{"type": "text", "title": "My Panel"},
		Name:      "My Library Panel",
		UID:       "lib-uid-abc",
	}
	entry := domain.WithNested[*models.LibraryElementDTO]{
		Entity:     &dto,
		NestedPath: "General/SubFolder",
	}

	got := WithNestedToCreateLibraryElement(entry)

	require.NotNil(t, got)
	assert.Equal(t, "folder-uid-123", got.FolderUID)
	assert.Equal(t, int64(1), got.Kind)
	assert.Equal(t, dto.Model, got.Model)
	assert.Equal(t, "My Library Panel", got.Name)
	assert.Equal(t, "lib-uid-abc", got.UID)
}

func TestWithNestedToCreateLibraryElement_EmptyFields(t *testing.T) {
	dto := &models.LibraryElementDTO{}
	entry := domain.WithNested[*models.LibraryElementDTO]{
		Entity:     &dto,
		NestedPath: "",
	}

	got := WithNestedToCreateLibraryElement(entry)

	require.NotNil(t, got)
	assert.Empty(t, got.FolderUID)
	assert.Zero(t, got.Kind)
	assert.Nil(t, got.Model)
	assert.Empty(t, got.Name)
	assert.Empty(t, got.UID)
}

func TestWithNestedToCreateLibraryElement_DoesNotMutateSource(t *testing.T) {
	dto := &models.LibraryElementDTO{
		Name: "Original Name",
		UID:  "original-uid",
	}
	entry := domain.WithNested[*models.LibraryElementDTO]{
		Entity: &dto,
	}

	got := WithNestedToCreateLibraryElement(entry)
	require.NotNil(t, got)

	// Mutate the result and verify the source DTO is unaffected.
	got.Name = "Modified Name"
	got.UID = "modified-uid"

	assert.Equal(t, "Original Name", dto.Name)
	assert.Equal(t, "original-uid", dto.UID)
}
