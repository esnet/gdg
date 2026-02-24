package ptr

//func ValOf[T any](value *T) T {
//	return ValueOrDefault(t, )
//	if value == nil {
//		var a T
//		return a
//	}
//	return *value
//}

// ValueOrDefault returns the value of a pointer if it is not nil,
// or the provided fallback value otherwise.
func ValueOrDefault[T any](v *T, fallback T) T {
	if v == nil {
		return fallback
	}
	return *v
}
