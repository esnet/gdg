package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/gosimple/slug"
	"github.com/netsage-project/grafana-dashboard-manager/config"
	"github.com/netsage-project/sdk"
)

//ListDataSources: list all the currently configured datasources
func ListDataSources(client *sdk.Client, folderFilters []string) []sdk.Datasource {

	ctx := context.Background()
	ds, err := client.GetAllDatasources(ctx)
	if err != nil {
		panic(err)
	}

	return ds
}

//ImportDataSources: will read in all the configured datasources.
//NOTE: credentials cannot be retrieved and need to be set via configuration
func ImportDataSources(client *sdk.Client, conf config.Provider) []string {
	var (
		datasources []sdk.Datasource
		dsPacked    []byte
		meta        sdk.BoardProperties
		err         error
		dataFiles   []string
	)
	datasources = ListDataSources(client, nil)
	for _, ds := range datasources {
		if dsPacked, err = json.MarshalIndent(ds, "", "	"); err != nil {
			fmt.Fprintf(os.Stderr, "%s for %s\n", err, ds.Name)
			continue
		}
		dsPath := buildDataSourcePath(conf, slug.Make(ds.Name))
		if err = ioutil.WriteFile(dsPath, dsPacked, os.FileMode(int(0666))); err != nil {
			fmt.Fprintf(os.Stderr, "%s for %s\n", err, meta.Slug)
		} else {
			dataFiles = append(dataFiles, dsPath)
		}
	}
	return dataFiles
}

//Removes all current datasources
func DeleteAllDataSources(client *sdk.Client) []string {
	ctx := context.Background()
	var ds []string = make([]string, 0)
	items := ListDataSources(client, nil)
	for _, item := range items {
		client.DeleteDatasource(ctx, item.ID)
		ds = append(ds, item.Name)
	}
	return ds
}

//ExportDataSources: exports all datasources to grafana using the credentials configured in config file.
func ExportDataSources(client *sdk.Client, folderFilters []string, query string, conf config.Provider) []string {
	var datasources []sdk.Datasource
	var status sdk.StatusMessage
	var exported []string = make([]string, 0)

	ctx := context.Background()
	filesInDir, err := ioutil.ReadDir(getResourcePath(conf, "ds"))
	datasources = ListDataSources(client, nil)

	var rawDS []byte
	if err != nil {
		fmt.Fprint(os.Stderr, err)
	}
	for _, file := range filesInDir {
		fileLocation := fmt.Sprintf("%s/%s", getResourcePath(conf, "ds"), file.Name())
		if strings.HasSuffix(file.Name(), ".json") {
			if rawDS, err = ioutil.ReadFile(fileLocation); err != nil {
				fmt.Fprint(os.Stderr, err)
				continue
			}
			var newDS sdk.Datasource

			if err = json.Unmarshal(rawDS, &newDS); err != nil {
				fmt.Fprint(os.Stderr, err)
				continue
			}
			user := config.GetGrafanaConfig().Datasource.User
			var secureData map[string]string = make(map[string]string, 0)
			newDS.BasicAuthUser = &user
			secureData["basicAuthPassword"] = config.GetGrafanaConfig().Datasource.Password

			newDS.SecureJSONData = secureData

			for _, existingDS := range datasources {
				if existingDS.Name == newDS.Name {
					if status, err = client.DeleteDatasource(ctx, existingDS.ID); err != nil {
						fmt.Fprintf(os.Stderr, "error on deleting datasource %s with %s", newDS.Name, err)
					}
					break
				}
			}
			if status, err = client.CreateDatasource(ctx, newDS); err != nil {
				fmt.Fprintf(os.Stderr, "error on importing datasource %s with %s (%s)", newDS.Name, err, *status.Message)
			} else {
				exported = append(exported, fileLocation)
			}

		}
	}
	return exported
}
