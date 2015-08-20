package memcache

// Cache is an interface that the different memory cache implementations will implement
type Cache interface {

	// Add adds a CacheItem to the cache, it can be retrieved using Get and passing in the same key
	Add(key string, val CacheItem) error

	// Get retrieves an item from the cache if its present
	//
	// If item is present then the item, true is returned. Otherwise, nil, false
	Get(key string) (CacheItem, bool)

	// Remove removes an item from the cache
	Remove(key string)
}

// CacheItem represents a single item in the cache
type CacheItem interface {

	// Size returns the size in memory of the item
	//
	// This can be used by the cache to keep track of the total size
	Size() int
}