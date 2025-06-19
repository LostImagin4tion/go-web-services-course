package main

func filter[T any](s *[]T, predicate func(T) bool) []T {
	var filtered = make([]T, 0)
	for _, item := range *s {
		if predicate(item) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

func mapped[T any, R any](s *[]T, mapper func(int, T) R) []R {
	var mapped = make([]R, len(*s))
	for i, item := range *s {
		mapped[i] = mapper(i, item)
	}
	return mapped
}

func copied[T any](s *[]T) []T {
	var newSlice = make([]T, len(*s))
	copy(newSlice, *s)
	return newSlice
}
