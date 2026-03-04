package tools

import (
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
