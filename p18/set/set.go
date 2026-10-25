package set

type Set[T comparable] struct {
	data map[T]struct{}
}

func NewSet[T comparable]() *Set[T] {
	return &Set[T]{
		data: make(map[T]struct{}),
	}
}

func (s *Set[T]) Add(value T) {
	s.data[value] = struct{}{}
}

func (s *Set[T]) Remove(value T) {
	delete(s.data, value)
}

func (s *Set[T]) Contains(value T) bool {
	_, exists := s.data[value]
	return exists
}

func (s *Set[T]) Len() int {
	return len(s.data)
}

func (s *Set[T]) Values() []T {
	values := make([]T, 0, len(s.data))
	for value := range s.data {
		values = append(values, value)
	}
	return values
}

func Union[T comparable](a, b *Set[T]) *Set[T] {
	result := NewSet[T]()
	for value := range a.data {
		result.Add(value)
	}
	for value := range b.data {
		result.Add(value)
	}
	return result
}

func Intersection[T comparable](a, b *Set[T]) *Set[T] {
	result := NewSet[T]()
	for value := range a.data {
		if b.Contains(value) {
			result.Add(value)
		}
	}
	return result
}

func Difference[T comparable](a, b *Set[T]) *Set[T] {
	result := NewSet[T]()
	for value := range a.data {
		if !b.Contains(value) {
			result.Add(value)
		}
	}
	return result
}
