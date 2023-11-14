package main

import (
	"fmt"
	assets "github.com/esnet/gdg/config"
	"github.com/esnet/gdg/internal/config"
	appconfig "github.com/esnet/gdg/internal/log"
	"github.com/esnet/gdg/internal/templating"
	"github.com/jedib0t/go-pretty/v6/table"
	flag "github.com/spf13/pflag"
	"gopkg.in/yaml.v3"
	"log"
	"log/slog"
	"os"
)

func main() {
	//Using pflag over corba for now, as this should be a simple enough CLI tool
	var cfgName = flag.StringP("config", "c", "importer.yml", "GDG Configuration file override.")
	var tmpCfgName = flag.StringP("ct", "", "templates.yml", "GDG Template configuration file override.")
	var showTemplateCfg = flag.BoolP("show-config", "", false, "Will display the current template configuration")
	var listTemplates = flag.BoolP("list-templates", "", false, "List all current templates")
	var templateName = flag.StringP("template", "t", "", "Specify template name, optional.  Default is to operate on all configured templates that are found.")
	flag.Parse()
	defaultConfiguration, err := assets.GetFile("importer-example.yml")
	if err != nil {
		slog.Warn("unable to load default configuration, no fallback")
	}

	config.InitConfig(*cfgName, defaultConfiguration)
	config.InitTemplateConfig(*tmpCfgName)
	cfg := config.Config()
	appconfig.InitializeAppLogger(os.Stdout, os.Stderr, cfg.IsDebug())

	if *showTemplateCfg {
		data, err := yaml.Marshal(cfg.GetTemplateConfig())
		if err != nil {
			slog.Error("unable to load template configuration")
		}
		slog.Info(fmt.Sprintf("Configuration\n%s", string(data)))
		return
	}
	slog.Info("Context is set to: ", slog.String("context", cfg.GetGDGConfig().ContextName))
	template := templating.NewTemplate()

	if *listTemplates {
		templates := template.ListTemplates()
		for ndx, t := range templates {
			slog.Info(fmt.Sprintf("%d: %s", ndx+1, t))
		}

		return
	}

	payload, err := template.Generate(*templateName)
	if err != nil {
		log.Fatal("Failed to generate templates", slog.Any("err", err))
	}

	tableObj := table.NewWriter()
	tableObj.SetOutputMirror(os.Stdout)
	tableObj.SetStyle(table.StyleLight)

	tableObj.AppendHeader(table.Row{"Template Name", "Output"})
	for key, val := range payload {
		for _, file := range val {
			tableObj.AppendRow(table.Row{key, file})
		}
	}

	tableObj.Render()
}
