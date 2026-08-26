package couplingmanifest

func swap[T any](values []T) []T {
	result := append([]T(nil), values...)
	if len(result) > 1 {
		result[0], result[1] = result[1], result[0]
	}
	return result
}
