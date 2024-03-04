package templating

import (
	"fmt"
	"github.com/Masterminds/sprig/v3"
	"github.com/esnet/gdg/internal/config"
	"github.com/esnet/gdg/internal/service"
	"github.com/esnet/gdg/internal/tools"
	"log/slog"
	"os"
	"strings"
	"text/template"
)

type Templating interface {
	Generate(templateName string) (map[string][]string, error)
	ListTemplates() []string
}

type templateImpl struct {
}

func NewTemplate() Templating {
	return &templateImpl{}
}

var fns = template.FuncMap{
	"ToSlug": service.GetSlug,
	"QuotedStringJoin": func(arr []interface{}) string {
		result := ""
		for ndx, item := range arr {
			if len(arr)-1 == ndx {
				result += fmt.Sprintf("\"%v\"", item)
			} else {
				result += fmt.Sprintf("\"%v\",", item)
			}
		}

		return result
	},
}

func (t *templateImpl) ListTemplates() []string {
	cfg := config.Config()
	var result []string
	entities := cfg.GetTemplateConfig().Entities.Dashboards
	for _, entry := range entities {
		result = append(result, entry.TemplateName)
	}

	return result
}

func (t *templateImpl) Generate(templateName string) (map[string][]string, error) {
	result := make(map[string][]string)
	//Remove extension if included
	templateName = strings.ReplaceAll(templateName, ".go.tmpl", "")
	cfg := config.Config()
	var entities []config.TemplateDashboards
	entities = cfg.GetTemplateConfig().Entities.Dashboards
	if templateName != "" {
		entity, ok := cfg.GetTemplateConfig().GetTemplate(templateName)
		if ok {
			entities = append(entities, *entity)
		}
	}
	for _, entity := range entities {
		result[entity.TemplateName] = make([]string, 0)
		slog.Info("Processing template:", slog.String("template", entity.TemplateName))
		tmplPath := cfg.GetDefaultGrafanaConfig().GetPath(config.TemplatesResource)
		fileLocation := fmt.Sprintf("%s/%s.go.tmpl", tmplPath, entity.TemplateName)
		_, err := os.Stat(fileLocation)
		if err != nil {
			slog.Error("Processing template, file could not be found", "template", entity.TemplateName, "file", fileLocation)
			slog.Warn("Continuing to process remaining templates")
			continue
		}
		templateData, err := os.ReadFile(fileLocation)
		if err != nil {
			slog.Error("unable to open file", slog.Any("file", fileLocation))
			slog.Warn("Continuing to process remaining templates")
			continue
		}
		for _, outputEntity := range entity.DashboardEntities {
			grafana := cfg.GetDefaultGrafanaConfig()
			slog.Debug("Creating a new template",
				slog.String("folder", outputEntity.Folder),
				slog.String("orgName", outputEntity.OrganizationName),
				slog.Any("data", outputEntity.TemplateData),
			)
			grafana.OrganizationName = outputEntity.OrganizationName
			outputPath := service.BuildResourceFolder(outputEntity.Folder, config.DashboardResource)
			//Merge two maps.
			tmpl, err := template.New("").Funcs(fns).Parse(string(templateData))
			if err != nil {
				slog.Error("unable to parse template")
			}

			//Create new file.
			tools.CreateDestinationPath(outputPath)
			dashboardName := entity.TemplateName
			if outputEntity.DashboardName != "" {
				dashboardName = service.GetSlug(outputEntity.DashboardName)
			}
			f, err := os.Create(fmt.Sprintf("%s/%s.json", outputPath, dashboardName))
			if err != nil {
				slog.Error("unable to create file: ", slog.Any("err", err))
				result[entity.TemplateName] = append(result[entity.TemplateName], err.Error())
				continue
			}
			slog.Debug("Writing data to destination", "output", f.Name())
			result[entity.TemplateName] = append(result[entity.TemplateName], f.Name())
			defer func() {
				err = f.Close()
				if err != nil {
					slog.Warn("failed to close template file", "filename", f.Name())
				}
			}()

			err = tmpl.Execute(f, outputEntity.TemplateData) // merge.
			if err != nil {
				slog.Error("execute", "err", err)
				result[entity.TemplateName] = append(result[entity.TemplateName], err.Error())
				continue
			}
		}
	}
	return result, nil

}

func init() {
	for key, value := range sprig.TxtFuncMap() {
		fns[key] = value
	}
}
