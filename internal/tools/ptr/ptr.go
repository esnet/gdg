package ptr

// Of returns a pointer to the given value.
//
// This is a convenience function to create a pointer from a value.
//
// Example:
//
//	p := ptr.Of(5)
//
// will return a pointer to the value 5.
func Of[T any](value T) *T {
	return &value
}
