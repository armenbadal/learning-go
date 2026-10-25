package cache

import (
	"container/list"
	"fmt"
)

type LRUCache[K comparable, V any] struct {
	capacity int
	links    map[K]*list.Element
	data     *list.List
}

type entry[K comparable, V any] struct {
	key   K
	value V
}

func NewLRUCache[K comparable, V any](capacity int) (*LRUCache[K, V], error) {
	if capacity <= 0 {
		return nil, fmt.Errorf("capacity must be greater than 0")
	}

	return &LRUCache[K, V]{
		capacity: capacity,
		links:    make(map[K]*list.Element),
		data:     list.New(),
	}, nil
}

func (c *LRUCache[K, V]) Get(key K) (V, bool) {
	if elem, ok := c.links[key]; ok {
		c.data.MoveToFront(elem)
		return elem.Value.(V), true
	}

	var zero V
	return zero, false
}

func (c *LRUCache[K, V]) Put(key K, value V) {
	if elem, ok := c.links[key]; ok {
		c.data.MoveToFront(elem)
		elem.Value = entry[K, V]{key: key, value: value}
		return
	}

	if c.Len() >= c.capacity {
		c.data.Remove(c.data.Back())
		delete(c.links, c.data.Back().Value.(entry[K, V]).key)
	}

	c.data.PushFront(entry[K, V]{key: key, value: value})
	c.links[key] = c.data.Front()
}

func (c *LRUCache[K, V]) Len() int {
	return len(c.links)
}
