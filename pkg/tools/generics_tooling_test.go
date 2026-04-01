package tools

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
)

// --- test types ---

type address struct {
	Street string
	City   string
}

type person struct {
	Name    string
	Age     int
	Tags    []string
	Address address
}

// --- tests ---

// TestDeepCopy_SimpleStruct verifies that a basic struct is copied correctly.
func TestDeepCopy_SimpleStruct(t *testing.T) {
	original := person{Name: "Alice", Age: 30}

	clone, err := DeepCopy(original)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if clone == nil {
		t.Fatal("expected non-nil clone")
	}
	if clone.Name != original.Name || clone.Age != original.Age {
		t.Errorf("clone mismatch: got %+v, want %+v", *clone, original)
	}
}

// TestDeepCopy_Independence verifies that mutating the clone does not affect the original.
func TestDeepCopy_Independence(t *testing.T) {
	original := person{
		Name: "Bob",
		Tags: []string{"go", "developer"},
	}

	clone, err := DeepCopy(original)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Mutate the clone.
	clone.Name = "Charlie"
	clone.Tags[0] = "rust"

	if original.Name != "Bob" {
		t.Errorf("original.Name was modified: got %q", original.Name)
	}
	if original.Tags[0] != "go" {
		t.Errorf("original.Tags[0] was modified: got %q", original.Tags[0])
	}
}

// TestDeepCopy_NestedStruct verifies that nested structs are deep-copied.
func TestDeepCopy_NestedStruct(t *testing.T) {
	original := person{
		Name:    "Dana",
		Address: address{Street: "123 Main St", City: "Springfield"},
	}

	clone, err := DeepCopy(original)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Mutate nested field on the clone.
	clone.Address.City = "Shelbyville"

	if original.Address.City != "Springfield" {
		t.Errorf("original.Address.City was modified: got %q", original.Address.City)
	}
}

// TestDeepCopy_Slice verifies that a standalone slice is deep-copied correctly.
func TestDeepCopy_Slice(t *testing.T) {
	original := []int{1, 2, 3, 4, 5}

	clone, err := DeepCopy(original)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(*clone) != len(original) {
		t.Fatalf("length mismatch: got %d, want %d", len(*clone), len(original))
	}

	// Mutate the clone and verify the original is unchanged.
	(*clone)[0] = 99
	if original[0] != 1 {
		t.Errorf("original slice was modified: got %d", original[0])
	}
}

// TestDeepCopy_Map verifies that a map is deep-copied correctly.
func TestDeepCopy_Map(t *testing.T) {
	original := map[string]int{"a": 1, "b": 2}

	clone, err := DeepCopy(original)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	(*clone)["a"] = 99
	if original["a"] != 1 {
		t.Errorf("original map was modified: got %d", original["a"])
	}
}

// TestDeepCopy_Pointer verifies that a pointer value is handled correctly.
func TestDeepCopy_Pointer(t *testing.T) {
	val := 42
	original := &val

	clone, err := DeepCopy(original)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if **clone != 42 {
		t.Errorf("expected 42, got %d", **clone)
	}

	// Mutate through the clone's pointer and verify original is unchanged.
	**clone = 100
	if val != 42 {
		t.Errorf("original value was modified: got %d", val)
	}
}

// TestDeepCopy_Primitive verifies that primitive types (string, int) work.
func TestDeepCopy_Primitive(t *testing.T) {
	t.Run("string", func(t *testing.T) {
		original := "hello"
		clone, err := DeepCopy(original)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if *clone != original {
			t.Errorf("got %q, want %q", *clone, original)
		}
	})

	t.Run("int", func(t *testing.T) {
		original := 123
		clone, err := DeepCopy(original)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if *clone != original {
			t.Errorf("got %d, want %d", *clone, original)
		}
	})
}

// TestDeepCopy_NonMarshalable verifies that a type that cannot be JSON-marshalled returns an error.
func TestDeepCopy_NonMarshalable(t *testing.T) {
	// Channels cannot be marshalled to JSON.
	ch := make(chan int)
	_, err := DeepCopy(ch)
	if err == nil {
		t.Error("expected an error for non-marshalable type, got nil")
	}
}

// ── ReverseLookUp ─────────────────────────────────────────────────────────────

// TestReverseLookUp_StringToString verifies a basic string-to-string reversal.
func TestReverseLookUp_StringToString(t *testing.T) {
	m := map[string]string{"a": "x", "b": "y", "c": "z"}
	rev := ReverseLookUp(m)
	if len(rev) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(rev))
	}
	for wantVal, wantKey := range map[string]string{"x": "a", "y": "b", "z": "c"} {
		if got := rev[wantVal]; got != wantKey {
			t.Errorf("rev[%q]: got %q, want %q", wantVal, got, wantKey)
		}
	}
}

// TestReverseLookUp_StringToInt verifies that a string→int map is reversed to int→string.
func TestReverseLookUp_StringToInt(t *testing.T) {
	m := map[string]int{"one": 1, "two": 2, "three": 3}
	rev := ReverseLookUp(m)
	if len(rev) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(rev))
	}
	cases := map[int]string{1: "one", 2: "two", 3: "three"}
	for k, want := range cases {
		if got := rev[k]; got != want {
			t.Errorf("rev[%d]: got %q, want %q", k, got, want)
		}
	}
}

// TestReverseLookUp_IntToString verifies that an int→string map is reversed to string→int.
func TestReverseLookUp_IntToString(t *testing.T) {
	m := map[int]string{10: "ten", 20: "twenty", 30: "thirty"}
	rev := ReverseLookUp(m)
	cases := map[string]int{"ten": 10, "twenty": 20, "thirty": 30}
	for k, want := range cases {
		if got := rev[k]; got != want {
			t.Errorf("rev[%q]: got %d, want %d", k, got, want)
		}
	}
}

// TestReverseLookUp_EmptyMap verifies that an empty map produces an empty reverse map.
func TestReverseLookUp_EmptyMap(t *testing.T) {
	m := map[string]string{}
	rev := ReverseLookUp(m)
	if len(rev) != 0 {
		t.Errorf("expected empty reverse map, got %d entries", len(rev))
	}
}

// TestReverseLookUp_SingleEntry verifies a map with exactly one entry.
func TestReverseLookUp_SingleEntry(t *testing.T) {
	m := map[int]string{42: "answer"}
	rev := ReverseLookUp(m)
	if len(rev) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(rev))
	}
	if rev["answer"] != 42 {
		t.Errorf(`rev["answer"]: got %d, want 42`, rev["answer"])
	}
}

// TestReverseLookUp_DuplicateValues_OneEntryPerUniqueValue verifies that when two
// keys share the same value the reverse map contains exactly one entry per unique
// value (the last key encountered in range wins, which is non-deterministic, but
// the length invariant is deterministic).
func TestReverseLookUp_DuplicateValues_OneEntryPerUniqueValue(t *testing.T) {
	m := map[string]string{
		"key1": "shared",
		"key2": "shared",
		"key3": "unique",
	}
	rev := ReverseLookUp(m)
	// Two distinct values → two entries, regardless of which key won the "shared" slot.
	if len(rev) != 2 {
		t.Errorf("expected 2 entries (one per unique value), got %d", len(rev))
	}
	if rev["unique"] != "key3" {
		t.Errorf(`rev["unique"]: got %q, want "key3"`, rev["unique"])
	}
	got := rev["shared"]
	if got != "key1" && got != "key2" {
		t.Errorf(`rev["shared"]: got %q, want "key1" or "key2"`, got)
	}
}

// TestReverseLookUp_DoubleReverseRestoresOriginal verifies that reversing twice
// returns the original map, provided all values are unique (no key collisions).
func TestReverseLookUp_DoubleReverseRestoresOriginal(t *testing.T) {
	original := map[string]int{"alpha": 1, "beta": 2, "gamma": 3}
	restored := ReverseLookUp(ReverseLookUp(original))
	if len(restored) != len(original) {
		t.Fatalf("length mismatch: got %d, want %d", len(restored), len(original))
	}
	for k, want := range original {
		if got := restored[k]; got != want {
			t.Errorf("restored[%q]: got %d, want %d", k, got, want)
		}
	}
}

// ── CreateDestinationPath ─────────────────────────────────────────────────────

// resetSyncMap replaces the package-level syncMap with a fresh one so each test
// starts from a clean slate.  Must only be called from tests that run sequentially.
func resetSyncMap(t *testing.T) {
	t.Helper()
	syncMap = new(sync.Map)
	t.Cleanup(func() { syncMap = new(sync.Map) })
}

// TestCreateDestinationPath_CreatesDirectory verifies that the target path is
// created by MkdirAll even when clearOutput is false.
func TestCreateDestinationPath_CreatesDirectory(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "output")
	resetSyncMap(t)

	CreateDestinationPath(dir, false, target)

	if _, err := os.Stat(target); os.IsNotExist(err) {
		t.Errorf("expected directory %q to be created, but it does not exist", target)
	}
}

// TestCreateDestinationPath_CreatesNestedPath verifies that MkdirAll creates
// all intermediate directories in a deep path.
func TestCreateDestinationPath_CreatesNestedPath(t *testing.T) {
	dir := t.TempDir()
	nested := filepath.Join(dir, "a", "b", "c")
	resetSyncMap(t)

	CreateDestinationPath(dir, false, nested)

	if _, err := os.Stat(nested); os.IsNotExist(err) {
		t.Errorf("expected nested directory %q to be created, but it does not exist", nested)
	}
}

// TestCreateDestinationPath_ClearOutputFalse_PreservesFolderName verifies that
// when clearOutput is false the folderName directory and its contents are left
// untouched.
func TestCreateDestinationPath_ClearOutputFalse_PreservesFolderName(t *testing.T) {
	dir := t.TempDir()
	folder := filepath.Join(dir, "backup")
	if err := os.MkdirAll(folder, 0o750); err != nil {
		t.Fatal(err)
	}
	sentinel := filepath.Join(folder, "keep.txt")
	if err := os.WriteFile(sentinel, []byte("keep"), 0o600); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(dir, "output")
	resetSyncMap(t)

	CreateDestinationPath(folder, false, target)

	if _, err := os.Stat(sentinel); os.IsNotExist(err) {
		t.Error("folderName contents should be preserved when clearOutput=false")
	}
}

// TestCreateDestinationPath_ClearOutputTrue_RemovesFolderName verifies that the
// first call with clearOutput=true removes the folderName directory.
func TestCreateDestinationPath_ClearOutputTrue_RemovesFolderName(t *testing.T) {
	dir := t.TempDir()
	folder := filepath.Join(dir, "backup")
	if err := os.MkdirAll(folder, 0o750); err != nil {
		t.Fatal(err)
	}
	sentinel := filepath.Join(folder, "old-data.txt")
	if err := os.WriteFile(sentinel, []byte("old"), 0o600); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(dir, "output")
	resetSyncMap(t)

	CreateDestinationPath(folder, true, target)

	if _, err := os.Stat(folder); !os.IsNotExist(err) {
		t.Errorf("folderName %q should have been removed when clearOutput=true", folder)
	}
}

// TestCreateDestinationPath_ClearOutputTrue_OnlyRemovesOnce verifies that the
// syncMap guard prevents a second call with the same folderName from removing
// data that was written after the first call.
func TestCreateDestinationPath_ClearOutputTrue_OnlyRemovesOnce(t *testing.T) {
	dir := t.TempDir()
	folder := filepath.Join(dir, "backup")
	if err := os.MkdirAll(folder, 0o750); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(dir, "output")
	resetSyncMap(t)

	// First call: removes folder and creates target.
	CreateDestinationPath(folder, true, target)

	// Recreate the folder with new content to verify the second call leaves it.
	if err := os.MkdirAll(folder, 0o750); err != nil {
		t.Fatal(err)
	}
	sentinel := filepath.Join(folder, "second-run.txt")
	if err := os.WriteFile(sentinel, []byte("new"), 0o600); err != nil {
		t.Fatal(err)
	}

	// Second call with the same folderName: syncMap already has the entry → skip removal.
	CreateDestinationPath(folder, true, target)

	if _, err := os.Stat(sentinel); os.IsNotExist(err) {
		t.Error("second call with the same folderName must not remove it again")
	}
}

// TestCreateDestinationPath_ClearOutputTrue_TargetCreatedAfterRemoval verifies
// that the target directory (v) is still created even though clearOutput removed
// the folderName.
func TestCreateDestinationPath_ClearOutputTrue_TargetCreatedAfterRemoval(t *testing.T) {
	dir := t.TempDir()
	folder := filepath.Join(dir, "backup")
	if err := os.MkdirAll(folder, 0o750); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(dir, "fresh-output")
	resetSyncMap(t)

	CreateDestinationPath(folder, true, target)

	if _, err := os.Stat(target); os.IsNotExist(err) {
		t.Errorf("target directory %q should be created even when clearOutput removed folderName", target)
	}
}
