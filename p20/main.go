package main

import "elru/cache"

func main() {
	cache, _ := cache.NewLRUCache[string, int](2)

	cache.Put("a", 10)
	cache.Put("b", 20)

	cache.Get("a") // a-ն դառնում է ամենավերջին օգտագործվածը
	cache.Put("c", 30)

	// b-ն հեռացվում է, քանի որ այն ամենահին օգտագործվածն է
}
