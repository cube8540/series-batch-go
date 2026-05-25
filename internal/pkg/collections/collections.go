package collections

func Find[E any](arr []E, predicate func(E) bool) int {
	for i, v := range arr {
		if predicate(v) {
			return i
		}
	}
	return -1
}

func MapAllValues[K comparable, V any](m map[K]V) []V {
	var values []V
	for _, v := range m {
		values = append(values, v)
	}
	return values
}

func ToAnySlices[T any](slice []T) []any {
	var result []any
	for _, v := range slice {
		result = append(result, v)
	}
	return result
}
