package service

import (
	"encoding/json"
	"github.com/esnet/gdg/internal/config"
	"github.com/esnet/grafana-swagger-api-golang/goclient/client/legacy_alerts_notification_channels"
	"github.com/esnet/grafana-swagger-api-golang/goclient/models"
	"strings"

	"github.com/gosimple/slug"
	log "github.com/sirupsen/logrus"
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
	params := legacy_alerts_notification_channels.NewGetAlertNotificationChannelsParams()
	channels, err := s.client.LegacyAlertsNotificationChannels.GetAlertNotificationChannels(params, s.getAuth())
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
			log.Errorf("error marshalling %s to json with %s", an.Name, err)
			continue
		}
		anPath := buildResourcePath(slug.Make(an.Name), config.AlertNotificationResource)
		if err = s.storage.WriteFile(anPath, anPacked); err != nil {
			log.Errorf("error writing %s to file with %s", slug.Make(an.Name), err)
		} else {
			dataFiles = append(dataFiles, anPath)
		}
	}
	return dataFiles
}

// Removes all current alert notification channels
func (s *DashNGoImpl) DeleteAllAlertNotifications() []string {
	var an []string = make([]string, 0)
	items := s.ListAlertNotifications()
	for _, item := range items {
		params := legacy_alerts_notification_channels.NewDeleteAlertNotificationChannelParams()
		params.NotificationChannelID = item.ID
		_, err := s.client.LegacyAlertsNotificationChannels.DeleteAlertNotificationChannel(params, s.getAuth())
		if err != nil {
			log.Error("Failed to delete notification")
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

			var newAlertNotification models.CreateAlertNotificationCommand
			if err = json.Unmarshal(raw, &newAlertNotification); err != nil {
				log.Errorf("error unmarshalling json with %s", err)
				continue
			}

			for _, existing := range alertnotifications {
				if existing.Name == newAlertNotification.Name {
					dp := legacy_alerts_notification_channels.NewDeleteAlertNotificationChannelByUIDParams()
					dp.NotificationChannelUID = existing.UID
					if _, err := s.client.LegacyAlertsNotificationChannels.DeleteAlertNotificationChannelByUID(dp, s.getAuth()); err != nil {
						log.Errorf("error on deleting datasource %s with %s", newAlertNotification.Name, err)
					}
					break
				}
			}

			params := legacy_alerts_notification_channels.NewCreateAlertNotificationChannelParams()
			params.Body = &newAlertNotification
			if _, err = s.client.LegacyAlertsNotificationChannels.CreateAlertNotificationChannel(params, s.getAuth()); err != nil {
				log.Errorf("error on importing datasource %s with %s", newAlertNotification.Name, err)
				continue
			}
			exported = append(exported, file)
		}
	}
	return exported
}
