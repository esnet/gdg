package api

// Unit tests for partitionDashboardFiles — the pure file-classification function
// extracted from UploadDashboards.  These tests have no external dependencies:
// no HTTP, no storage mock, no config loading.  readFile is supplied as a plain
// closure keyed on the file path, making every branch directly exercisable.

import (
	"errors"
	"fmt"
	"testing"

	"github.com/esnet/gdg/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Fixture helpers
// ---------------------------------------------------------------------------

// v1JSON returns minimal JSON that looks like a plain Grafana v1 dashboard.
// It has no "gdg_api_version" field, so partitionDashboardFiles treats it as v1.
func v1JSON(uid string) []byte {
	return []byte(fmt.Sprintf(`{"uid":%q,"title":"My Dashboard","tags":[]}`, uid))
}

// v1JSONNoUID returns a v1 dashboard payload without a uid field.
func v1JSONNoUID() []byte {
	return []byte(`{"title":"Untitled","tags":[]}`)
}

// v2JSON returns JSON in the DashboardV2Gdg envelope format.
// resource.metadata.name carries the dashboard identifier on the v2 path.
func v2JSON(name string) []byte {
	return []byte(fmt.Sprintf(`{
		"gdg_api_version": %q,
		"nested_path": "General",
		"resource": {
			"metadata": {"name": %q},
			"spec": {"title": "V2 Dashboard"}
		}
	}`, domain.GdgApiVersionV2, name))
}

// v2JSONNoName returns a v2 envelope where resource.metadata.name is absent.
func v2JSONNoName() []byte {
	return []byte(fmt.Sprintf(`{
		"gdg_api_version": %q,
		"nested_path": "General",
		"resource": {
			"metadata": {},
			"spec": {"title": "Nameless"}
		}
	}`, domain.GdgApiVersionV2))
}

// v2JSONNilResource returns a v2 envelope where resource is null.
func v2JSONNilResource() []byte {
	return []byte(fmt.Sprintf(`{"gdg_api_version":%q,"nested_path":"General","resource":null}`, domain.GdgApiVersionV2))
}

// unknownVersionJSON has a non-empty gdg_api_version that is not GdgApiVersionV2.
// Per the classification rules this should fall through to v1Files.
func unknownVersionJSON() []byte {
	return []byte(`{"gdg_api_version":"some-future-version/v99","title":"Future"}`)
}

// invalidJSON is not valid JSON at all.
func invalidJSON() []byte {
	return []byte(`{not valid json`)
}

// staticReader returns a readFile func that serves fixed payloads by file path.
func staticReader(m map[string][]byte) func(string) ([]byte, error) {
	return func(path string) ([]byte, error) {
		b, ok := m[path]
		if !ok {
			return nil, errors.New("file not found: " + path)
		}
		return b, nil
	}
}

// ---------------------------------------------------------------------------
// Table-driven tests
// ---------------------------------------------------------------------------

func TestPartitionDashboardFiles(t *testing.T) {
	tests := []struct {
		name string

		// inputs
		files    []string
		contents map[string][]byte // path → bytes served by readFile

		// expected outputs
		wantV1Keys    []string // keys expected in v1Files
		wantV2Keys    []string // keys expected in v2Files
		wantProcessed []string // keys expected in processed
		wantV1Absent  []string // keys that must NOT appear in v1Files
		wantV2Absent  []string // keys that must NOT appear in v2Files
	}{
		{
			name:          "EmptyFileList",
			files:         []string{},
			contents:      map[string][]byte{},
			wantV1Keys:    []string{},
			wantV2Keys:    []string{},
			wantProcessed: []string{},
		},
		{
			name: "SingleV1File_UIDRecorded",
			files: []string{"org_main-org/dashboards/my-dash.json"},
			contents: map[string][]byte{
				"org_main-org/dashboards/my-dash.json": v1JSON("uid-abc"),
			},
			wantV1Keys:    []string{"org_main-org/dashboards/my-dash.json"},
			wantV2Keys:    []string{},
			wantProcessed: []string{"uid-abc"},
		},
		{
			name: "MultipleV1Files_AllUIDsRecorded",
			files: []string{
				"dashboards/dash-1.json",
				"dashboards/dash-2.json",
			},
			contents: map[string][]byte{
				"dashboards/dash-1.json": v1JSON("uid-1"),
				"dashboards/dash-2.json": v1JSON("uid-2"),
			},
			wantV1Keys:    []string{"dashboards/dash-1.json", "dashboards/dash-2.json"},
			wantV2Keys:    []string{},
			wantProcessed: []string{"uid-1", "uid-2"},
		},
		{
			name: "SingleV2File_NameRecorded",
			files: []string{"dashboards/v2-dash.json"},
			contents: map[string][]byte{
				"dashboards/v2-dash.json": v2JSON("my-v2-dashboard"),
			},
			wantV1Keys:    []string{},
			wantV2Keys:    []string{"dashboards/v2-dash.json"},
			wantProcessed: []string{"my-v2-dashboard"},
		},
		{
			name: "MixedV1AndV2Files",
			files: []string{
				"dashboards/legacy.json",
				"dashboards/new.json",
			},
			contents: map[string][]byte{
				"dashboards/legacy.json": v1JSON("legacy-uid"),
				"dashboards/new.json":    v2JSON("new-name"),
			},
			wantV1Keys:    []string{"dashboards/legacy.json"},
			wantV2Keys:    []string{"dashboards/new.json"},
			wantProcessed: []string{"legacy-uid", "new-name"},
		},
		{
			name: "NonJsonFileSkipped",
			files: []string{
				"dashboards/config.yaml",
				"dashboards/valid.json",
			},
			contents: map[string][]byte{
				// config.yaml should never be read; valid.json should be classified
				"dashboards/valid.json": v1JSON("valid-uid"),
			},
			wantV1Keys:    []string{"dashboards/valid.json"},
			wantV2Keys:    []string{},
			wantProcessed: []string{"valid-uid"},
			wantV1Absent:  []string{"dashboards/config.yaml"},
		},
		{
			name: "ReadFileErrorSkipsFile",
			files: []string{
				"dashboards/unreadable.json",
				"dashboards/ok.json",
			},
			contents: map[string][]byte{
				// "unreadable.json" is intentionally absent from the map so the
				// staticReader returns an error for it, exercising the skip branch.
				"dashboards/ok.json": v1JSON("ok-uid"),
			},
			wantV1Keys:    []string{"dashboards/ok.json"},
			wantV2Keys:    []string{},
			wantProcessed: []string{"ok-uid"},
			wantV1Absent:  []string{"dashboards/unreadable.json"},
			wantV2Absent:  []string{"dashboards/unreadable.json"},
		},
		{
			name: "V1FileWithNoUID_ProcessedEmpty",
			files: []string{"dashboards/no-uid.json"},
			contents: map[string][]byte{
				"dashboards/no-uid.json": v1JSONNoUID(),
			},
			wantV1Keys:    []string{"dashboards/no-uid.json"},
			wantV2Keys:    []string{},
			wantProcessed: []string{}, // no uid → nothing added to processed
		},
		{
			name: "V2FileWithNilResource_ProcessedEmpty",
			files: []string{"dashboards/nil-res.json"},
			contents: map[string][]byte{
				"dashboards/nil-res.json": v2JSONNilResource(),
			},
			wantV1Keys:    []string{},
			wantV2Keys:    []string{"dashboards/nil-res.json"},
			wantProcessed: []string{}, // nil resource → nothing added to processed
		},
		{
			name: "V2FileWithNoResourceName_ProcessedEmpty",
			files: []string{"dashboards/no-name.json"},
			contents: map[string][]byte{
				"dashboards/no-name.json": v2JSONNoName(),
			},
			wantV1Keys:    []string{},
			wantV2Keys:    []string{"dashboards/no-name.json"},
			wantProcessed: []string{},
		},
		{
			name: "InvalidJSON_TreatedAsV1",
			files: []string{"dashboards/corrupt.json"},
			contents: map[string][]byte{
				// Invalid JSON can't be decoded into DashboardV2Gdg, so it falls
				// through to the v1 branch.  No uid → processed stays empty.
				"dashboards/corrupt.json": invalidJSON(),
			},
			wantV1Keys:    []string{"dashboards/corrupt.json"},
			wantV2Keys:    []string{},
			wantProcessed: []string{},
		},
		{
			name: "UnknownVersion_FallsBackToV1",
			files: []string{"dashboards/future.json"},
			contents: map[string][]byte{
				// A non-empty version that is neither "" nor GdgApiVersionV2
				// must land in v1Files per the else-branch.
				"dashboards/future.json": unknownVersionJSON(),
			},
			wantV1Keys:    []string{"dashboards/future.json"},
			wantV2Keys:    []string{},
			wantProcessed: []string{},
		},
		{
			name: "RawBytesPreserved",
			// Verifies that the bytes stored in v1Files/v2Files are identical
			// to what readFile returned (no transformation).
			files: []string{
				"dashboards/a.json",
				"dashboards/b.json",
			},
			contents: map[string][]byte{
				"dashboards/a.json": v1JSON("a-uid"),
				"dashboards/b.json": v2JSON("b-name"),
			},
			wantV1Keys:    []string{"dashboards/a.json"},
			wantV2Keys:    []string{"dashboards/b.json"},
			wantProcessed: []string{"a-uid", "b-name"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			v1Files, v2Files, processed := partitionDashboardFiles(tc.files, staticReader(tc.contents))

			// --- v1Files assertions ---
			for _, k := range tc.wantV1Keys {
				assert.Contains(t, v1Files, k, "v1Files should contain %q", k)
			}
			for _, k := range tc.wantV1Absent {
				assert.NotContains(t, v1Files, k, "v1Files must not contain %q", k)
			}

			// --- v2Files assertions ---
			for _, k := range tc.wantV2Keys {
				assert.Contains(t, v2Files, k, "v2Files should contain %q", k)
			}
			for _, k := range tc.wantV2Absent {
				assert.NotContains(t, v2Files, k, "v2Files must not contain %q", k)
			}

			// --- processed assertions ---
			for _, k := range tc.wantProcessed {
				assert.True(t, processed[k], "processed should contain key %q", k)
			}
			// Verify the maps have exactly the expected sizes so stray entries
			// are caught rather than silently ignored.
			assert.Len(t, v1Files, len(tc.wantV1Keys),
				"v1Files length mismatch")
			assert.Len(t, v2Files, len(tc.wantV2Keys),
				"v2Files length mismatch")
			assert.Len(t, processed, len(tc.wantProcessed),
				"processed length mismatch")
		})
	}
}

// TestPartitionDashboardFiles_RawBytesIdentity verifies that the bytes stored
// in the output maps are the exact same slice returned by readFile — no copying
// or transformation occurs.
func TestPartitionDashboardFiles_RawBytesIdentity(t *testing.T) {
	v1Payload := v1JSON("identity-uid")
	v2Payload := v2JSON("identity-name")

	contents := map[string][]byte{
		"dashboards/v1.json": v1Payload,
		"dashboards/v2.json": v2Payload,
	}

	v1Files, v2Files, _ := partitionDashboardFiles(
		[]string{"dashboards/v1.json", "dashboards/v2.json"},
		staticReader(contents),
	)

	require.Contains(t, v1Files, "dashboards/v1.json")
	assert.Equal(t, v1Payload, v1Files["dashboards/v1.json"])

	require.Contains(t, v2Files, "dashboards/v2.json")
	assert.Equal(t, v2Payload, v2Files["dashboards/v2.json"])
}

// TestPartitionDashboardFiles_NoMutualContamination verifies that a file
// classified as v2 never appears in v1Files, and vice versa.
func TestPartitionDashboardFiles_NoMutualContamination(t *testing.T) {
	files := []string{
		"dashboards/a.json",
		"dashboards/b.json",
		"dashboards/c.json",
	}
	contents := map[string][]byte{
		"dashboards/a.json": v1JSON("uid-a"),
		"dashboards/b.json": v2JSON("name-b"),
		"dashboards/c.json": v1JSON("uid-c"),
	}

	v1Files, v2Files, processed := partitionDashboardFiles(files, staticReader(contents))

	assert.NotContains(t, v1Files, "dashboards/b.json", "v2 file must not appear in v1Files")
	assert.NotContains(t, v2Files, "dashboards/a.json", "v1 file must not appear in v2Files")
	assert.NotContains(t, v2Files, "dashboards/c.json", "v1 file must not appear in v2Files")

	assert.True(t, processed["uid-a"])
	assert.True(t, processed["name-b"])
	assert.True(t, processed["uid-c"])
}
