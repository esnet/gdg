package templating

import (
	"fmt"
	"log/slog"
	"maps"
	"os"
	"strings"
	"text/template"

	"github.com/esnet/gdg/internal/config/domain"
	resourceTypes "github.com/esnet/gdg/pkg/config/domain"

	"github.com/esnet/gdg/internal/tools"

	"github.com/Masterminds/sprig/v3"
	"github.com/esnet/gdg/internal/service"
)

type Templating interface {
	Generate(templateName string) (map[string][]string, error)
	ListTemplates() []string
}

type templateImpl struct {
	cfg    *domain.TemplatingConfig
	gdgCfg *domain.GrafanaConfig
}

func NewTemplate(cfg *domain.TemplatingConfig, gdgCfg *domain.GrafanaConfig) Templating {
	return &templateImpl{
		cfg:    cfg,
		gdgCfg: gdgCfg,
	}
}

var fns = template.FuncMap{
	"ToSlug": service.GetSlug,
	"QuotedStringJoin": func(arr []any) string {
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
	var result []string
	entities := t.cfg.Entities.Dashboards
	for _, entry := range entities {
		result = append(result, entry.TemplateName)
	}

	return result
}

func (t *templateImpl) Generate(templateName string) (map[string][]string, error) {
	result := make(map[string][]string)
	// Remove extension if included
	var entities []domain.TemplateDashboards
	templateName = strings.ReplaceAll(templateName, ".go.tmpl", "")
	templateEntity, ok := t.cfg.GetTemplate(templateName)
	if ok {
		entities = append(entities, *templateEntity)
	} else {
		entities = t.cfg.Entities.Dashboards
	}
	for _, entity := range entities {
		result[entity.TemplateName] = make([]string, 0)
		slog.Info("Processing template:", slog.String("template", entity.TemplateName))
		tmplPath := t.gdgCfg.GetPath(resourceTypes.TemplatesResource, t.gdgCfg.GetOrganizationName())
		fileLocation := fmt.Sprintf("%s/%s.go.tmpl", tmplPath, entity.TemplateName)
		_, err := os.Stat(fileLocation)
		if err != nil {
			slog.Error("Processing template, file could not be found", "template", entity.TemplateName, "file", fileLocation)
			slog.Warn("Continuing to process remaining templates")
			continue
		}
		templateData, err := os.ReadFile(fileLocation) // #nosec G304
		if err != nil {
			slog.Error("unable to open file", slog.Any("file", fileLocation))
			slog.Warn("Continuing to process remaining templates")
			continue
		}
		for _, outputEntity := range entity.DashboardEntities {
			grafana := t.gdgCfg
			slog.Debug("Creating a new template",
				slog.String("folder", outputEntity.Folder),
				slog.String("orgName", outputEntity.OrganizationName),
				slog.Any("data", outputEntity.TemplateData),
			)
			grafana.OrganizationName = outputEntity.OrganizationName
			outputPath := service.BuildResourceFolder(t.gdgCfg, outputEntity.Folder, resourceTypes.DashboardResource, true, false)
			// Merge two maps.
			tmpl, tmplErr := template.New("").Funcs(fns).Parse(string(templateData))
			if tmplErr != nil {
				slog.Error("unable to parse template", slog.Any("err", err))
				continue
			}

			// Create new file.
			tools.CreateDestinationPath("", false, outputPath)
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
	maps.Copy(fns, sprig.TxtFuncMap())
}
