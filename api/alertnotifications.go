package api

import (
	"context"
	"encoding/json"
	"github.com/esnet/gdg/config"
	"os"
	"strings"

	"github.com/gosimple/slug"
	"github.com/grafana-tools/sdk"
	log "github.com/sirupsen/logrus"
)

//ListAlertNotifications: list all currently configured notification channels
func (s *DashNGoImpl) ListAlertNotifications() []sdk.AlertNotification {
	ctx := context.Background()
	ans, err := s.client.GetAllAlertNotifications(ctx)
	if err != nil {
		log.Panic(err)
	}
	return ans
}

//ImportAlertNotifications: will read in all the configured alert notification channels.
func (s *DashNGoImpl) ImportAlertNotifications() []string {
	var (
		alertnotifications []sdk.AlertNotification
		anPacked           []byte
		meta               sdk.BoardProperties
		err                error
		dataFiles          []string
	)
	alertnotifications = s.ListAlertNotifications()
	for _, an := range alertnotifications {
		if anPacked, err = json.Marshal(an); err != nil {
			log.Errorf("error marshalling %s to json with %s", an.Name, err)
			continue
		}
		anPath := buildResourcePath(slug.Make(an.Name), config.AlertNotificationResource)
		if err = s.storage.WriteFile(anPath, anPacked, os.FileMode(int(0666))); err != nil {
			log.Errorf("error writing %s to file with %s", meta.Slug, err)
		} else {
			dataFiles = append(dataFiles, anPath)
		}
	}
	return dataFiles
}

//Removes all current alert notification channels
func (s *DashNGoImpl) DeleteAllAlertNotifications() []string {
	ctx := context.Background()
	var an []string = make([]string, 0)
	items := s.ListAlertNotifications()
	for _, item := range items {
		err := s.client.DeleteAlertNotificationID(ctx, uint(item.ID))
		if err != nil {
			log.Error("Failed to delete notification")
			continue
		}
		an = append(an, item.Name)
	}
	return an
}

//ExportAlertNotifications: exports all alert notification channels to grafana.
//NOTE: credentials will be missing and need to be set manually after export
//TODO implement configuring sensitive fields for different kinds of alert notification channels
func (s *DashNGoImpl) ExportAlertNotifications() []string {
	var (
		alertnotifications []sdk.AlertNotification
		exported           []string
		filesInDir         []string
		err                error
	)

	ctx := context.Background()
	dirPath := getResourcePath(config.AlertNotificationResource)
	filesInDir, err = s.storage.FindAllFiles(dirPath, true)
	if err != nil {
		log.WithError(err).Fatalf("Unable to find Alert data in Storage System %s", s.storage.Name())
	}
	alertnotifications = s.ListAlertNotifications()

	var raw []byte
	for _, file := range filesInDir {
		if strings.HasSuffix(file, ".json") {
			if raw, err = s.storage.ReadFile(file); err != nil {
				log.Errorf("error reading file %s with %s", file, err)
				continue
			}

			var new sdk.AlertNotification
			if err = json.Unmarshal(raw, &new); err != nil {
				log.Errorf("error unmarshalling json with %s", err)
				continue
			}

			for _, existing := range alertnotifications {
				if existing.Name == new.Name {
					if err = s.client.DeleteAlertNotificationID(ctx, uint(existing.ID)); err != nil {
						log.Errorf("error on deleting datasource %s with %s", new.Name, err)
					}
					break
				}
			}

			if _, err = s.client.CreateAlertNotification(ctx, new); err != nil {
				log.Errorf("error on importing datasource %s with %s", new.Name, err)
				continue
			}
			exported = append(exported, file)
		}
	}
	return exported
}
