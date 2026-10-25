package collections

func Filter[T any](items []T, predicate func(T) bool) []T {
	result := make([]T, 0)
	for _, item := range items {
		if predicate(item) {
			result = append(result, item)
		}
	}
	return result
}

func Map[T any, U any](items []T, transform func(T) U) []U {
	result := make([]U, len(items))
	for i, item := range items {
		result[i] = transform(item)
	}
	return result
}

func Unique[T comparable](items []T) []T {
	seen := make(map[T]bool)
	for _, item := range items {
		seen[item] = true
	}
	result := make([]T, 0, len(seen))
	for item := range seen {
		result = append(result, item)
	}
	return result
}

func GroupBy[T any, K comparable](items []T, keyFn func(T) K) map[K][]T {
	result := make(map[K][]T)
	for _, item := range items {
		key := keyFn(item)
		if _, exists := result[key]; !exists {
			result[key] = make([]T, 0)
		}
		result[key] = append(result[key], item)
	}
	return result
}

func Reduce[T any, U any](items []T, initial U, reducer func(U, T) U) U {
	result := initial
	for _, item := range items {
		result = reducer(result, item)
	}
	return result
}

func ForEach[T any](items []T, action func(T)) {
	for _, item := range items {
		action(item)
	}
}
