package utils

// Map transforms a slice of type T to a slice of type U.
func Map[T any, U any](input []T, transform func(T) U) []U {
	result := make([]U, len(input))
	for i, v := range input {
		result[i] = transform(v)
	}
	return result
}
