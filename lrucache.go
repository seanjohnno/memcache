package memcache

import (
	"errors"
	"sync"
	"fmt"
)

const (
	// ErrorExceedsMaxSize is the error returned by Add if the item is too big for the cache 
	ErrorExceedsMaxSize = "Exceeds max size, can't store"
)

// ------------------------------------------------------------------------------------------------------------------------
// Creation functions
// ------------------------------------------------------------------------------------------------------------------------

// CreateLRUCache creates and returns a 'Last Recently Used' implementation of Cache
//
// LRU keeps items added and accessed most recently in preference to older items. Older meaning, last added / accessed
func CreateLRUCache(maxsize int) (Cache) {
	return &lruCache { keyValMap: make(map[string]*lruCacheItem), maxSize: maxsize, mutex: sync.Mutex { } }
}

// ------------------------------------------------------------------------------------------------------------------------
// Struct: lruCacheItem (not exported)
// ------------------------------------------------------------------------------------------------------------------------

// lruCacheItem represents a single cache item
type lruCacheItem struct {

	// cacheItem is the underlying item - stored so we can find its size
	cacheItem CacheItem

	// key is the key we'd use in Get(key) to retrieve the item
	key string

	// prev is the previous item in the linked-list, nil if we're the head
	prev *lruCacheItem

	// next is the next item in the linked-list, nil if we're the tail
	next *lruCacheItem
}

// Remove removes this item from the lruCache and handles all clearup
//
// It pairs sibling nodes and points head/tail elsewhere if it was either. It also removes itself from the hash and alters 
// the cache size
func (this *lruCacheItem) Remove(cache *lruCache) {
	// Join up left and right nodes (or point them at nil if heads and tails)
	if this.prev != nil {
		this.prev.next = this.next
	}
	if this.next != nil {
		this.next.prev = this.prev
	}

	// Point head / tail at new node if this was either
	if this == cache.head {
		cache.head = this.next
	}
	if this == cache.tail {
		cache.tail = this.prev
	}

	// Remove size
	cache.curSize -= this.cacheItem.Size()

	// Remove from map
	delete(cache.keyValMap, this.key)
}

// Add adds this item to the head of the cache and alters the cache state accordigly
//
// It sets itself as the head and links the previous head and itself together. It also adds itself to the hash and alters
// the cache size
func (this *lruCacheItem) Add(cache *lruCache) {
	// If we're the only element then set head and tail
	if cache.head == nil {
		cache.head = this
		cache.tail = this

	// Otherwise, this is the new head
	} else {
		cache.head.prev = this
		this.next = cache.head
		cache.head = this
	}

	// Add size to cache
	cache.curSize += this.cacheItem.Size()

	// Add to map
	cache.keyValMap[this.key] = this
}

// ------------------------------------------------------------------------------------------------------------------------
// Struct: lruCache (not exported)
// ------------------------------------------------------------------------------------------------------------------------

// lruCache is a used to implement the LRU cache implementation 
//
// It keeps a hash and a linked-list of CacheItems. The hash is for fast access and the linked-list is used to drop items 
// off the end
type lruCache struct {

	// keyValMap is the map of key(string) to value(CacheItem)
	keyValMap map[string]*lruCacheItem
	
	// head is the head of the linkedlist
	head *lruCacheItem

	// tail is the tail of the linkedlist
	tail *lruCacheItem

	// maxSize holds the maximum size of the cache
	maxSize int

	// curSize holds the current size of the cache
	curSize int

	// mutex is used to synchronize cache as it can be accessed by 
	mutex sync.Mutex
}

// ------------------------------------------------------------------------------------------------------------------------
// Cache Implementation
// ------------------------------------------------------------------------------------------------------------------------

// Add adds a CacheItem to the cache, it can be retrieved using Get and passing in the same key
//
// If the item already exists its moved from its current place in the linked-list to the head
// If the item doesn't curretly exist then its added to the head. This item will add to the current size of the cache. If
// the current size > max size then tail items are removed until it falls under max size
func (this *lruCache) Add(k string, v CacheItem) error {

	// Lock method so hash ad linked-list can be accessed safely from multiple go-routines. Unlock when func returns
	this.mutex.Lock()
	defer this.mutex.Unlock()

	// If we already contain item then remove from linked-list (value may be different)
	if item, present := this.keyValMap[k]; present {
		// Removes from position in linked-list
		item.Remove(this)
		
		// Values are the same so we can just move to the start of the array
		if v == item.cacheItem {
			item.Add(this)
			return nil
		}
	}

	// Can't store if it already exceeds max size
	if v.Size() > this.maxSize {
		return errors.New(ErrorExceedsMaxSize)
	}

	// Remove tail items until we're under max size
	for this.curSize + v.Size() > this.maxSize {
		this.tail.Remove(this)
	}

	// Create item
	lruItem := &lruCacheItem { cacheItem: v, key: k }
	lruItem.Add(this)
	return nil
}

// Get retrieves an item from the cache if its present. Also, because its been accessed its moved to the head of the queue
//
// If item is present then the item, true is returned. Otherwise, nil, false
func (this *lruCache) Get(key string) (CacheItem, bool) {

	// Lock method so hash ad linked-list can be accessed safely from multiple go-routines. Unlock when func returns
	this.mutex.Lock()
	defer this.mutex.Unlock()

	// See if the cache contains the item
	if item, containsKey := this.keyValMap[key]; containsKey {
		item.Remove(this)
		item.Add(this)
		
		return item.cacheItem, containsKey
	}
	return nil, false
}

// Remove removes an item from the cache
func (this *lruCache) Remove(key string) {
	// Lock method so hash ad linked-list can be accessed safely from multiple go-routines. Unlock when func returns
	this.mutex.Lock()
	defer this.mutex.Unlock()

		// Check if item is present in cache
	lruCacheItem, present := this.keyValMap[key]
	if present {
		lruCacheItem.Remove(this)
	}
}