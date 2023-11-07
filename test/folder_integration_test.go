package test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"log/slog"
)

func TestFolderCRUD(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	apiClient, _, cleanup := initTest(t, nil)
	defer cleanup()
	slog.Info("Exporting all folders")
	apiClient.UploadFolders(nil)
	slog.Info("Listing all Folders")
	folders := apiClient.ListFolder(nil)
	assert.Equal(t, len(folders), 2)
	var firstDsItem = folders[0]
	assert.Equal(t, firstDsItem.Title, "Ignored")
	var secondDsItem = folders[1]
	assert.Equal(t, secondDsItem.Title, "Other")
	//Import Folders
	slog.Info("Importing folders")
	list := apiClient.DownloadFolders(nil)
	assert.Equal(t, len(list), len(folders))
	slog.Info("Deleting Folders")
	deleteList := apiClient.DeleteAllFolders(nil)
	assert.Equal(t, len(deleteList), len(folders))
	slog.Info("List Folders again")
	folders = apiClient.ListFolder(nil)
	assert.Equal(t, len(folders), 0)
}
