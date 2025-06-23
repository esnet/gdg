package service

import (
	"log"
	"log/slog"
	"path/filepath"
	"slices"
	"strings"

	"github.com/esnet/gdg/internal/config/domain"

	"github.com/esnet/gdg/internal/config"
	"github.com/esnet/gdg/internal/service/filters"
	"github.com/esnet/gdg/internal/service/types"
	"github.com/grafana/dashboard-linter/lint"
	"github.com/zeitlinger/conflate"
)

func (s *DashNGoImpl) LintDashboards(req types.LintRequest) []string {
	var rawBoard []byte
	dashboardPath := config.Config().GetDefaultGrafanaConfig().GetPath(domain.DashboardResource, s.grafanaConf.GetOrganizationName())
	filesInDir, err := s.storage.FindAllFiles(dashboardPath, true)
	if err != nil {
		log.Fatalf("unable to find any files to export from storage engine, err: %v", err)
	}
	filterReq := NewDashboardFilter(req.FolderName, req.DashboardSlug, "")
	validFolders, err := filterReq.GetExpectedStringSlice(filters.FolderFilter)
	if err != nil {
		log.Fatalf("unable to get expected folders from filter request, err: %v", err)
	}
	for _, file := range filesInDir {
		baseFile := filepath.Base(file)
		baseFile = strings.ReplaceAll(baseFile, ".json", "")

		if !strings.HasSuffix(file, ".json") {
			slog.Warn("Only json files are supported, skipping", "filename", file)
			continue
		}
		if req.DashboardSlug != "" && baseFile != req.DashboardSlug {
			slog.Debug("Skipping dashboard, does not match filter", slog.String("dashboard", req.DashboardSlug))
			continue
		}

		if rawBoard, err = s.storage.ReadFile(file); err != nil {
			slog.Warn("Unable to read file", "filename", file, "err", err)
			continue
		}
		if req.FolderName != "" {
			if !slices.Contains(validFolders, req.FolderName) && !config.Config().GetDefaultGrafanaConfig().GetDashboardSettings().IgnoreFilters {
				slog.Debug("Skipping file since it doesn't match any valid folders", "filename", file)
				continue
			}
		}

		dashboard, err := lint.NewDashboard(rawBoard)
		if err != nil {
			slog.Error("failed to parse dashboard", slog.Any("err", err))
			continue
		}
		lintConfigFlag := strings.ReplaceAll(file, ".json", ".lint")
		cfgLint := lint.NewConfigurationFile()
		if err := cfgLint.Load(lintConfigFlag); err != nil {
			slog.Error("Unable to load lintConfigFlag")
			continue
		}
		cfgLint.Verbose = req.VerboseFlag
		cfgLint.Autofix = req.AutoFix

		rules := lint.NewRuleSet()
		results, err := rules.Lint([]lint.Dashboard{dashboard})
		if err != nil {
			slog.Error("failed to lint dashboard", slog.Any("err", err))
			continue

		}
		if cfgLint.Autofix {
			changes := results.AutoFix(&dashboard)
			if changes > 0 {
				slog.Info("AutoFix possible")
				writeErr := s.writeLintedDashboard(dashboard, file, rawBoard)
				if writeErr != nil {
					slog.Error("unable to autofix linting issues for dashboard", slog.String("dashboard", file))
				}
			} else {
				slog.Error("AutoFix is not possible for dashboard.", slog.String("dashboard", file))
			}
		}

		slog.Info("Running Linter for Dashboard", slog.String("file", file))
		results.ReportByRule()

	}
	return nil
}

func (s *DashNGoImpl) writeLintedDashboard(dashboard lint.Dashboard, filename string, old []byte) error {
	newBytes, err := dashboard.Marshal()
	if err != nil {
		return err
	}
	c := conflate.New()
	err = c.AddData(old, newBytes)
	if err != nil {
		return err
	}
	b, err := c.MarshalJSON()
	if err != nil {
		return err
	}
	json := strings.ReplaceAll(string(b), "\"options\": null,", "\"options\": [],")
	return s.storage.WriteFile(filename, []byte(json))
}
