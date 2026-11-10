package main

import "fmt"

type Storage interface {
	Store(id uint, obj any) error
	Load(id uint) (any, error)
}

type MemoryStorage struct {
	data map[uint]any
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		data: make(map[uint]any),
	}
}

func (ms *MemoryStorage) Store(id uint, obj any) error {
	if _, exists := ms.data[id]; exists {
		return fmt.Errorf("'%d' ID exists", id)
	}

	ms.data[id] = obj
	return nil
}

func (ms *MemoryStorage) Load(id uint) (any, error) {
	obj, found := ms.data[id]
	if !found {
		return nil, fmt.Errorf("There is no object with ID=%d", id)
	}

	return obj, nil
}



func main() {
	ms := NewMemoryStorage()
	ms.Store(1, "ok")
	ms.Store(2, 123)
	ms.Store(3, 3.1415)

	s, _ := ms.Load(1)
	v := s.(string)
	println(s, v)
}
