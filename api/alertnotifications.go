package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/gosimple/slug"
	"github.com/grafana-tools/sdk"
	log "github.com/sirupsen/logrus"
)

//ListAlertNotifications: list all currently configured notification channels
func (s *DashNGoImpl) ListAlertNotifications(filter Filter) []sdk.AlertNotification {
	ctx := context.Background()
	an, err := s.client.GetAllAlertNotifications(ctx)
	if err != nil {
		panic(err)
	}
	result := make([]sdk.AlertNotification, 0)
	for _, item := range an {
		if filter.Validate(map[string]string{Name: GetSlug(item.Name)}) {
			result = append(result, item)
		}
	}

	return result
}

//ImportAlertNotifications: will read in all the configured alert notification channels.
//NOTE: sensitive fields cannot be retrieved and need to be set via configuration
func (s *DashNGoImpl) ImportAlertNotifications(filter Filter) []string {
	var (
		alertnotifications []sdk.AlertNotification
		anPacked           []byte
		meta               sdk.BoardProperties
		err                error
		dataFiles          []string
	)
	alertnotifications = s.ListAlertNotifications(filter)
	for _, an := range alertnotifications {
		if anPacked, err = json.Marshal(an); err != nil {
			log.Errorf("%s for %s\n", err, an.Name)
			continue
		}
		anPath := buildAlertNotificationPath(s.configRef, slug.Make(an.Name))
		if err = ioutil.WriteFile(anPath, anPacked, os.FileMode(int(0666))); err != nil {
			log.Errorf("%s for %s\n", err, meta.Slug)
		} else {
			dataFiles = append(dataFiles, anPath)
		}
	}
	return dataFiles
}

//Removes all current alert notification channels
func (s *DashNGoImpl) DeleteAllAlertNotifications(filter Filter) []string {
	ctx := context.Background()
	var an []string = make([]string, 0)
	items := s.ListAlertNotifications(filter)
	for _, item := range items {
		s.client.DeleteAlertNotificationID(ctx, uint(item.ID))
		an = append(an, item.Name)
	}
	return an
}

//ExportDataSources: exports all alert notification channels to grafana.
func (s *DashNGoImpl) ExportAlertNotifications(filter Filter) []string {
	var alertnotifications []sdk.AlertNotification
	var status int64
	var exported []string = make([]string, 0)

	ctx := context.Background()
	filesInDir, err := ioutil.ReadDir(getResourcePath(s.configRef, "an"))
	alertnotifications = s.ListAlertNotifications(filter)

	var rawAN []byte
	if err != nil {
		fmt.Fprint(os.Stderr, err)
	}
	for _, file := range filesInDir {
		fileLocation := filepath.Join(getResourcePath(s.configRef, "an"), file.Name())
		if strings.HasSuffix(file.Name(), ".json") {
			if rawAN, err = ioutil.ReadFile(fileLocation); err != nil {
				fmt.Fprint(os.Stderr, err)
				continue
			}
			var newAN sdk.AlertNotification

			if err = json.Unmarshal(rawAN, &newAN); err != nil {
				fmt.Fprint(os.Stderr, err)
				continue
			}

			if !filter.Validate(map[string]string{Name: GetSlug(newAN.Name)}) {
				continue
			}

			for _, existingAN := range alertnotifications {
				if existingAN.Name == newAN.Name {
					if err = s.client.DeleteAlertNotificationID(ctx, uint(existingAN.ID)); err != nil {
						log.Errorf("error on deleting datasource %s with %s", newAN.Name, err)
					}
					break
				}
			}
			if status, err = s.client.CreateAlertNotification(ctx, newAN); err != nil {
				log.Errorf("error on importing datasource %s with %s (status %s)", newAN.Name, err, status)
			} else {
				exported = append(exported, fileLocation)
			}

		}
	}
	return exported
}
