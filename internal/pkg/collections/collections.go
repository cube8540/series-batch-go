package collections

func Find[E any](arr []E, predicate func(E) bool) int {
	for i, v := range arr {
		if predicate(v) {
			return i
		}
	}
	return -1
}

func Map[K comparable, V any](arr []V, mapper func(V) K) map[K]V {
	m := make(map[K]V)
	for _, v := range arr {
		k := mapper(v)
		m[k] = v
	}
	return m
}

func MapToSlice[K comparable, V any, E any](m map[K]V, extract func(K, V) E) []E {
	var slice []E
	for k, v := range m {
		slice = append(slice, extract(k, v))
	}
	return slice
}
