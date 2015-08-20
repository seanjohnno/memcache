package memcache

import (
	"errors"
	"sync"
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
// Structs (not exported)
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

// ------------------------------------------------------------------------------------------------------------------------
// Cache Implementation
// ------------------------------------------------------------------------------------------------------------------------

// Add adds a CacheItem to the cache, it can be retrieved using Get and passing in the same key
//
// If the item already exists its moved from its current place in the linked-list to the head
// If the item doesn't curretly exist then its added to the head. This item will add to the current size of the cache. If
// the current size > max size then tail items are removed until it falls under max size
func (this *lruCache) Add(k string, v CacheItem) error {

	this.Remove(k)

	// Create item
	lruItem := &lruCacheItem { cacheItem: v, key: k }

	// Can't store if it already exceeds max size
	if v.Size() > this.maxSize {
		return errors.New(ErrorExceedsMaxSize)
	}

	// Lock method so hash ad linked-list can be accessed safely from multiple go-routines. Unlock when func returns
	this.mutex.Lock()
	defer this.mutex.Unlock()

	// Remove tail items until we're under max size
	for this.curSize + v.Size() > this.maxSize {
		// Need to remove from hash as well as 
		delete(this.keyValMap, this.tail.key)
		this.curSize -= this.tail.cacheItem.Size()

		newTail := this.tail.prev
		if newTail != nil {
			newTail.next = nil
			this.tail = newTail
		} else {
			this.head = nil
			this.tail = nil
		}
	}

	// Add to map (locking as maps aren't thread safe)
	this.keyValMap[k] = lruItem

	// If we're empty then head/tail are both new item
	if this.head == nil {
		this.head = lruItem
		this.tail = lruItem

	// New item is new head
	} else {
		lruItem.next = this.head
		this.head.prev = lruItem
		this.head = lruItem
	}

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
	val, containsKey := this.keyValMap[key]
	if containsKey {
		// If so we move to the front of the linked-list
		if this.head != val {
			val.prev.next = val.next

			if val.next != nil {
				val.next.prev = val.prev
			}

			val.prev = nil
			val.next = this.head
			this.head.prev = val
			this.head = val
		}
		return val.cacheItem, containsKey
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
		delete(this.keyValMap, key)

		this.curSize -= lruCacheItem.cacheItem.Size()

		// If the node is the head of the linked list
		if lruCacheItem == this.head {
			if this.head.next != nil {
				lruCacheItem.next.prev = nil
				this.head = lruCacheItem.next
			} else {
				this.head = nil
				this.tail = nil
			}

		// If the node is the tail of the linked list
		} else if lruCacheItem == this.tail {
			
			if this.tail.prev != nil {
				lruCacheItem.prev.next = nil
				this.tail = lruCacheItem.prev
			} else {
				this.head = nil
				this.tail = nil
			}

		// If the nodes in the middle then join the neighbours up
		} else {
			lruCacheItem.prev.next = lruCacheItem.next
			lruCacheItem.next.prev = lruCacheItem.prev
		}
	}
}