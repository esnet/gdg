package test

import (
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestFolderCRUD(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	apiClient, _, cleanup := initTest(t, nil)
	defer cleanup()
	log.Info("Exporting all folders")
	apiClient.UploadFolders(nil)
	log.Info("Listing all Folders")
	folders := apiClient.ListFolder(nil)
	assert.Equal(t, len(folders), 2)
	var firstDsItem = folders[0]
	assert.Equal(t, firstDsItem.Title, "Ignored")
	var secondDsItem = folders[1]
	assert.Equal(t, secondDsItem.Title, "Other")
	//Import Folders
	log.Info("Importing folders")
	list := apiClient.DownloadFolders(nil)
	assert.Equal(t, len(list), len(folders))
	log.Info("Deleting Folders")
	deleteList := apiClient.DeleteAllFolders(nil)
	assert.Equal(t, len(deleteList), len(folders))
	log.Info("List Folders again")
	folders = apiClient.ListFolder(nil)
	assert.Equal(t, len(folders), 0)
}
