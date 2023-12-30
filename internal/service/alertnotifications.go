package service

import (
	"encoding/json"
	"github.com/esnet/gdg/internal/config"
	"github.com/grafana/grafana-openapi-client-go/models"
	"log/slog"
	"strings"

	"github.com/gosimple/slug"
	"log"
)

// AlertNotificationsApi Contract definition
// Deprecated: Marked as Deprecated as of Grafana 9.0, Moving to ContactPoints is recommended
type AlertNotificationsApi interface {
	ListAlertNotifications() []*models.AlertNotification
	DownloadAlertNotifications() []string
	UploadAlertNotifications() []string
	DeleteAllAlertNotifications() []string
}

//ListAlertNotifications: list all currently configured notification channels

func (s *DashNGoImpl) ListAlertNotifications() []*models.AlertNotification {
	channels, err := s.GetClient().LegacyAlertsNotificationChannels.GetAlertNotificationChannels()
	if err != nil {
		log.Panic(err)
	}
	return channels.Payload
}

// ImportAlertNotifications: will read in all the configured alert notification channels.
func (s *DashNGoImpl) DownloadAlertNotifications() []string {
	var (
		alertnotifications []*models.AlertNotification
		anPacked           []byte
		err                error
		dataFiles          []string
	)
	alertnotifications = s.ListAlertNotifications()
	for _, an := range alertnotifications {
		if anPacked, err = json.Marshal(an); err != nil {
			slog.Error("error marshalling to json", "filename", an.Name, "err", err.Error())
			continue
		}
		anPath := buildResourcePath(slug.Make(an.Name), config.AlertNotificationResource)
		if err = s.storage.WriteFile(anPath, anPacked); err != nil {
			slog.Error("error writing to file", "filename", slug.Make(an.Name), "err", err.Error())
		} else {
			dataFiles = append(dataFiles, anPath)
		}
	}
	return dataFiles
}

// Removes all current alert notification channels
func (s *DashNGoImpl) DeleteAllAlertNotifications() []string {
	var an = make([]string, 0)
	items := s.ListAlertNotifications()
	for _, item := range items {
		_, err := s.GetClient().LegacyAlertsNotificationChannels.DeleteAlertNotificationChannel(item.ID)
		if err != nil {
			slog.Error("Failed to delete notification")
			continue
		}
		an = append(an, item.Name)
	}
	return an
}

// ExportAlertNotifications: exports all alert notification channels to grafana.
// NOTE: credentials will be missing and need to be set manually after export
// TODO implement configuring sensitive fields for different kinds of alert notification channels
func (s *DashNGoImpl) UploadAlertNotifications() []string {
	var (
		alertnotifications []*models.AlertNotification
		exported           []string
		filesInDir         []string
		err                error
	)

	dirPath := config.Config().GetDefaultGrafanaConfig().GetPath(config.AlertNotificationResource)
	filesInDir, err = s.storage.FindAllFiles(dirPath, true)
	if err != nil {
		log.Fatalf("Unable to find Alert data in Storage System %s, err: %s", s.storage.Name(), err.Error())
	}
	alertnotifications = s.ListAlertNotifications()

	var raw []byte
	for _, file := range filesInDir {
		if strings.HasSuffix(file, ".json") {
			if raw, err = s.storage.ReadFile(file); err != nil {
				slog.Error("error reading file", "file", file, "err", err)
				continue
			}

			var newAlertNotification models.CreateAlertNotificationCommand
			if err = json.Unmarshal(raw, &newAlertNotification); err != nil {
				slog.Error("error unmarshalling json", "err", err)
				continue
			}

			for _, existing := range alertnotifications {
				if existing.Name == newAlertNotification.Name {
					if _, err := s.GetClient().LegacyAlertsNotificationChannels.DeleteAlertNotificationChannelByUID(existing.UID); err != nil {
						slog.Error("error on deleting datasource", "datasource", newAlertNotification.Name, "err", err)
					}
					break
				}
			}

			if _, err = s.GetClient().LegacyAlertsNotificationChannels.CreateAlertNotificationChannel(&newAlertNotification); err != nil {
				slog.Error("error on importing datasource", "datasource", newAlertNotification.Name, "err", err)
				continue
			}
			exported = append(exported, file)
		}
	}
	return exported
}
