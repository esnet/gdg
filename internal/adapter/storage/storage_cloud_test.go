package storage

import (
	"os"
	"strings"
	"testing"
)

func TestGetMapValue(t *testing.T) {
	data := map[string]string{
		"key1": "value1",
		"key2": "value2",
	}

	t.Run("Key exists and value is not empty", func(t *testing.T) {
		result := getMapValue("key1", "default", func(s string) bool { return s == "" }, data)
		if result != "value1" {
			t.Errorf("Expected 'value1', got '%s'", result)
		}
	})

	t.Run("Key does not exist, should return default", func(t *testing.T) {
		result := getMapValue("key3", "default", func(s string) bool { return s == "" }, data)
		if result != "default" {
			t.Errorf("Expected 'default', got '%s'", result)
		}
	})

	t.Run("Key exists but value is empty, should return default", func(t *testing.T) {
		dataWithEmpty := map[string]string{
			"key1": "",
			"key2": "value2",
		}
		result := getMapValue("key1", "default", func(s string) bool { return s == "" }, dataWithEmpty)
		if result != "default" {
			t.Errorf("Expected 'default', got '%s'", result)
		}
	})

	t.Run("Using integer type", func(t *testing.T) {
		intData := map[int]int{
			1: 10,
			2: 20,
		}
		resultInt := getMapValue(1, 0, func(i int) bool { return i == 0 }, intData)
		if resultInt != 10 {
			t.Errorf("Expected 10, got %d", resultInt)
		}
	})

	t.Run(" Key does not exist in integer map, should return default", func(t *testing.T) {
		intData := map[int]int{
			1: 10,
			2: 20,
		}
		resultInt := getMapValue(3, 0, func(i int) bool { return i == 0 }, intData)
		if resultInt != 0 {
			t.Errorf("Expected 0, got %d", resultInt)
		}
	})
}

func TestGetMapValueOrEnvOverride(t *testing.T) {
	// Set up test data
	testData := map[string]string{
		"key1": "value1",
		"key2": "value2",
	}

	t.Run("Environment variable overrides map value", func(t *testing.T) {
		// Set environment variable
		os.Setenv("TEST_ENV_KEY", "env_override_value")
		defer os.Unsetenv("TEST_ENV_KEY")

		result := getMapValueOrEnvOverride("key1", "TEST_ENV_KEY", testData)
		if result != "env_override_value" {
			t.Errorf("Expected 'env_override_value', got '%s'", result)
		}
	})

	t.Run("Environment variable not set, returns map value", func(t *testing.T) {
		// Ensure environment variable is not set
		os.Unsetenv("UNSET_TEST_ENV_KEY")

		result := getMapValueOrEnvOverride("key1", "UNSET_TEST_ENV_KEY", testData)
		if result != "value1" {
			t.Errorf("Expected 'value1', got '%s'", result)
		}
	})

	t.Run("Environment variable is empty string, returns map value", func(t *testing.T) {
		// Set environment variable to empty string
		os.Setenv("EMPTY_ENV_KEY", "")
		defer os.Unsetenv("EMPTY_ENV_KEY")

		result := getMapValueOrEnvOverride("key2", "EMPTY_ENV_KEY", testData)
		if result != "value2" {
			t.Errorf("Expected 'value2', got '%s'", result)
		}
	})

	t.Run("Non-existent key returns empty string", func(t *testing.T) {
		result := getMapValueOrEnvOverride("nonexistent", "NONEXISTENT_ENV_KEY", testData)
		if result != "" {
			t.Errorf("Expected '', got '%s'", result)
		}
	})
}

// ---------------------------------------------------------------------------
// boolStrCheck
// ---------------------------------------------------------------------------

func TestBoolStrCheck(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		// truthy values
		{"true", true},
		{"True", true},
		{"TRUE", true},
		{"tRuE", true},
		{"1", true},
		// falsy values
		{"false", false},
		{"False", false},
		{"FALSE", false},
		{"0", false},
		{"", false},
		{"yes", false},
		{"on", false},
		{"2", false},
	}
	for _, tt := range tests {
		t.Run("input="+tt.input, func(t *testing.T) {
			got := boolStrCheck(tt.input)
			if got != tt.want {
				t.Errorf("boolStrCheck(%q): got %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// getCloudLocation
// ---------------------------------------------------------------------------

func newCloudStorage(prefix string) *CloudStorage {
	return &CloudStorage{
		BucketName:  "test-bucket",
		Prefix:      prefix,
		StorageName: "test",
	}
}

func TestGetCloudLocation_NilPrefix(t *testing.T) {
	s := newCloudStorage("<nil>")
	got := s.getCloudLocation("org_main-org/dashboards/my-dash.json")
	// "<nil>" prefix must be cleared; result should have no prefix injected.
	if strings.HasPrefix(got, "<nil>") {
		t.Errorf("expected prefix to be cleared, got %q", got)
	}
	// The file path itself should be preserved.
	if !strings.Contains(got, "my-dash.json") {
		t.Errorf("expected file name in result, got %q", got)
	}
}

func TestGetCloudLocation_EmptyPrefix(t *testing.T) {
	s := newCloudStorage("")
	fileName := "org_main-org/dashboards/my-dash.json"
	got := s.getCloudLocation(fileName)
	if got != fileName {
		t.Errorf("empty prefix: got %q, want %q", got, fileName)
	}
}

func TestGetCloudLocation_FileAlreadyContainsPrefix(t *testing.T) {
	// When the fileName already contains the prefix the function must return it unchanged.
	s := newCloudStorage("myprefix")
	fileName := "myprefix/org_main-org/dashboards/my-dash.json"
	got := s.getCloudLocation(fileName)
	if got != fileName {
		t.Errorf("already-prefixed: got %q, want %q", got, fileName)
	}
}

func TestGetCloudLocation_FileWithoutLeadingSlash(t *testing.T) {
	s := newCloudStorage("myprefix")
	fileName := "org_main-org/dashboards/my-dash.json"
	got := s.getCloudLocation(fileName)
	// Must start with the prefix and contain the file name.
	if !strings.HasPrefix(got, "myprefix") {
		t.Errorf("expected result to start with prefix, got %q", got)
	}
	if !strings.Contains(got, "my-dash.json") {
		t.Errorf("expected file name in result, got %q", got)
	}
}

func TestGetCloudLocation_FileWithLeadingSlash(t *testing.T) {
	s := newCloudStorage("myprefix")
	fileName := "/org_main-org/dashboards/my-dash.json"
	got := s.getCloudLocation(fileName)
	// Must combine prefix and file name without double slashes.
	if strings.Contains(got, "//") {
		t.Errorf("double slash detected in result %q", got)
	}
	if !strings.Contains(got, "my-dash.json") {
		t.Errorf("expected file name in result, got %q", got)
	}
}
