package memcache

import (
	"testing"
	"strconv"
	"fmt"
)

const (
	MaxSize = 100
)

func TestLRUCache(t *testing.T) {
	cache := CreateLRUCache(MaxSize)
	
	// Check cache is created correctly
	if cache == nil {
		t.Error("Cache is nil")
	}

	// Add 10 cacheitems of size 10, all should remain present
	for i := 0; i < 10; i++ {
		cache.Add(strconv.Itoa(i), &DummyCacheItem{DummySize: 10})
	}

	// Check all our keys remain
	// for i := 0; i < 10; i++ {
	// 	if _, present := cache.Get(strconv.Itoa(i)); !present {
	// 		t.Error(strconv.Itoa(i), "should be present")
	// 	}
	// }

	fmt.Println("\nAfter adds...\n")

	// 9 should be last accessed so front of queue, 0 should be last
	cache.Add("10", &DummyCacheItem{DummySize: 10})

	// We should have gne over max size so zero should have been knocked off the queue
	if _, present := cache.Get("0"); present {
		t.Error("0 should have been removed from the end of the queue")
	}

	// 1 should now be at the end of the queue, access to put it at the start
	cache.Get("1")
	
	// Add another and 2 not 1 should be knocked off the end of the queue
	cache.Add("11", &DummyCacheItem{DummySize: 10})

	// Check we have 1 but don't have 2
	if _, present := cache.Get("1"); !present {
		t.Error("1 should be present in the queue")
	}

	if _, present := cache.Get("2"); present {
		t.Error("2 should have been removed from the queue")
	}
}

type DummyCacheItem struct {
	DummySize int
}

func (this *DummyCacheItem) Size() int {
	return this.DummySize
}