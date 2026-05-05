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
