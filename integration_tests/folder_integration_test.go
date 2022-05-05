package integration_tests

import (
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestFolderCRUD(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	apiClient, _ := initTest(t)
	log.Info("Exporting all folders")
	apiClient.ExportFolder(nil)
	log.Info("Listing all Folders")
	folders := apiClient.ListFolder(nil)
	assert.Equal(t, len(folders), 1)
	var dsItem = folders[0]
	assert.Equal(t, dsItem.Title, "Other")
	//Import Folders
	log.Info("Importing folders")
	list := apiClient.ImportFolder(nil)
	assert.Equal(t, len(list), len(folders))
	log.Info("Deleting Folders")
	deleteList := apiClient.DeleteAllFolder(nil)
	assert.Equal(t, len(deleteList), len(folders))
	log.Info("List Folders again")
	folders = apiClient.ListFolder(nil)
	assert.Equal(t, len(folders), 0)
}
