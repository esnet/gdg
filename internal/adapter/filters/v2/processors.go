package v2

import (
	"context"
	"log/slog"
	"regexp"

	"github.com/esnet/gdg/internal/domain"
	"github.com/samber/lo"
)

// FolderQuoteRegExProcessor is a shared DataProcessor that strips single and double quotes
// from folder name strings before they are matched against regex patterns. It handles both
// a plain string value (from live API objects) and a []string slice (from expected value
// lists). Registered by any filter that performs folder-based regex validation.
var FolderQuoteRegExProcessor = domain.ProcessorEntity{
	Name: "folderQuoteRegEx",
	Processor: func(ctx context.Context, item any) (any, error) {
		quoteRegex := regexp.MustCompile("['\"]+")
		switch w := item.(type) {
		case string:
			slog.Debug("folder quote filter applied to string")
			return quoteRegex.ReplaceAllString(w, ""), nil
		case []string:
			slog.Debug("folder quote filter applied to []string")
			return lo.Map(w, func(i string, _ int) string {
				return quoteRegex.ReplaceAllString(i, "")
			}), nil
		}
		return item, nil
	},
}
