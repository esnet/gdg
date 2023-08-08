package service

import (
	"encoding/json"
	"errors"
	"github.com/esnet/gdg/internal/config"
	"github.com/esnet/grafana-swagger-api-golang/goclient/client/signed_in_user"
	"github.com/esnet/grafana-swagger-api-golang/goclient/models"
	"github.com/gosimple/slug"
	"os"
	"path/filepath"
	"strings"

	"github.com/esnet/grafana-swagger-api-golang/goclient/client/orgs"
	log "github.com/sirupsen/logrus"
)

// OrganizationsApi Contract definition
type OrganizationsApi interface {
	ListOrganizations() []*models.OrgDTO
	DownloadOrganizations() []string
	UploadOrganizations() []string
	SetOrganization(id int64) error
}

func (s *DashNGoImpl) getOrganization(id int64) (*models.OrgDetailsDTO, error) {
	params := orgs.NewGetOrgByIDParams()
	params.OrgID = id
	data, err := s.client.Orgs.GetOrgByID(params, s.getAuth())
	if err != nil {
		return nil, err
	}

	return data.GetPayload(), nil

}

func (s *DashNGoImpl) SetOrganization(id int64) error {
	//Removes Org filter
	if id == 0 {
		s.grafanaConf.Organization = ""
	} else {
		org, err := s.getOrganization(id)
		if err != nil {
			return errors.New("invalid org Id, org is not found")
		}
		s.grafanaConf.Organization = org.Name
	}

	return config.Config().SaveToDisk(false)
}

// ListOrganizations List all dashboards
func (s *DashNGoImpl) ListOrganizations() []*models.OrgDTO {
	orgList, err := s.client.Orgs.SearchOrgs(orgs.NewSearchOrgsParams(), s.getAdminAuth())
	if err != nil {
		log.WithError(err).Errorf("Unable to retrieve Organization List")
	}
	if s.grafanaConf.Organization != "" {
		var ID int64
		for _, org := range orgList.Payload {
			log.Errorf("%d %s\n", org.ID, org.Name)
			if org.Name == s.grafanaConf.Organization {
				ID = org.ID
			}
		}
		if ID > 0 {
			params := signed_in_user.NewUserSetUsingOrgParams()
			params.OrgID = ID
			status, err := s.client.SignedInUser.UserSetUsingOrg(params, s.getAuth())
			if err != nil {
				log.Fatalf("%s for %v\n", err, status)
			}
		}
	}

	orgList, err = s.client.Orgs.SearchOrgs(orgs.NewSearchOrgsParams(), s.getAuth())
	if err != nil {
		panic(err)
	}
	return orgList.GetPayload()
}

// DownloadOrganizations Download organizations
func (s *DashNGoImpl) DownloadOrganizations() []string {
	var (
		dsPacked  []byte
		err       error
		dataFiles []string
	)

	orgsListing := s.ListOrganizations()
	for _, organisation := range orgsListing {
		if dsPacked, err = json.MarshalIndent(organisation, "", "	"); err != nil {
			log.Errorf("%s for %s\n", err, organisation.Name)
			continue
		}
		dsPath := buildResourcePath(slug.Make(organisation.Name), config.OrganizationResource)
		if err = s.storage.WriteFile(dsPath, dsPacked, os.FileMode(int(0666))); err != nil {
			log.Errorf("%s for %s\n", err.Error(), slug.Make(organisation.Name))
		} else {
			dataFiles = append(dataFiles, dsPath)
		}
	}

	return dataFiles
}

func (s *DashNGoImpl) UploadOrganizations() []string {
	var (
		result    []string
		rawFolder []byte
	)
	filesInDir, err := s.storage.FindAllFiles(config.Config().GetDefaultGrafanaConfig().GetPath(config.OrganizationResource), false)
	if err != nil {
		log.WithError(err).Fatal("Failed to read folders imports")
	}
	orgListing := s.ListOrganizations()
	orgMap := map[string]bool{}
	for _, entry := range orgListing {
		orgMap[entry.Name] = true
	}

	for _, file := range filesInDir {
		fileLocation := filepath.Join(config.Config().GetDefaultGrafanaConfig().GetPath(config.OrganizationResource), file)
		if strings.HasSuffix(file, ".json") {
			if rawFolder, err = s.storage.ReadFile(fileLocation); err != nil {
				log.WithError(err).Errorf("failed to read file %s", fileLocation)
				continue
			}
		}
		var newOrg models.CreateOrgCommand
		if err = json.Unmarshal(rawFolder, &newOrg); err != nil {
			log.WithError(err).Warn("failed to unmarshall folder")
			continue
		}
		if _, ok := orgMap[newOrg.Name]; ok {
			log.Infof("Organizaiton %s already exists, skipping", newOrg.Name)
			continue
		}

		params := orgs.NewCreateOrgParams()
		params.Body = &newOrg
		_, err = s.client.Orgs.CreateOrg(params, s.getAdminAuth())
		if err != nil {
			log.Errorf("failed to create folder %s", newOrg.Name)
			continue
		}
		result = append(result, newOrg.Name)

	}
	return result
}
