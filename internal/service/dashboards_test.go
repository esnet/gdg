package service

import (
	"github.com/esnet/gdg/internal/config"
	"github.com/grafana/grafana-openapi-client-go/models"
	"github.com/stretchr/testify/assert"
	"testing"
)

func createDummyDashboard(dbName string, dbId int64, parent *models.Hit) *models.Hit {
	return &models.Hit{
		FolderID:    parent.ID,
		FolderTitle: parent.Title,
		FolderUID:   parent.UID,
		Title:       dbName,
		ID:          dbId,
		Type:        "dash-db",
	}
}

func createDummyFolder(folderTitle string, folderId int64, folderUid string, parent *models.Hit) *models.Hit {
	var parentId int64 = 0
	var parentTitle = ""
	var parentUID = ""
	if parent != nil {
		parentId = parent.ID
		parentTitle = parent.Title
		parentUID = parent.UID
	}
	return &models.Hit{
		FolderID:    parentId,
		FolderTitle: parentTitle,
		FolderUID:   parentUID,
		Title:       folderTitle,
		ID:          folderId,
		UID:         folderUid,
		Type:        "dash-db",
	}
}

func TestHitBehaviour(t *testing.T) {
	rootFolder := createDummyFolder("root", 21, "x21", nil)

	assert.Equal(t, "", rootFolder.FolderUID)
}

func TestBuildDashboardFileName(t *testing.T) {
	expectedResult := "test/data/org_main-org/dashboards/root/narf/zort.json"
	expectedResultOther := "test/data/org_main-org/dashboards/root/poit.json"

	rootFolder := createDummyFolder("root", 21, "x21", nil)
	subFolder := createDummyFolder("narf", 23, "x23", rootFolder)
	dummyDashboard := createDummyDashboard("ZORT", 42, subFolder)
	dummyDashboardOther := createDummyDashboard("POIT", 7, rootFolder)

	boardList := make([]*models.Hit, 0)
	boardList = append(boardList, rootFolder)
	boardList = append(boardList, dummyDashboardOther)
	boardList = append(boardList, subFolder)
	boardList = append(boardList, dummyDashboard)

	config.InitGdgConfig("testing.yml", "'")
	result := buildDashboardFileName(dummyDashboard, "zort", boardList)
	resultOther := buildDashboardFileName(dummyDashboardOther, "poit", boardList)

	assert.NotNil(t, result)
	assert.Equal(t, expectedResult, result)

	assert.NotNil(t, resultOther)
	assert.Equal(t, expectedResultOther, resultOther)
}
