## MemCache for golang

### Description

In-memory caching of objects using a map like interface. Interface can be found [here](https://github.com/seanjohnno/memcache/blob/master/memcache.go)

```
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
```

Currently the only implementation of the interface is [LRU Cache](https://github.com/seanjohnno/memcache/blob/master/lrucache.go), I'll add more as I go along...

### LRU Cache

The LRU implementation allows you to pick a max cache size. If an item is added or accessed it is placed or moved to the front of the queue. If the cache goes beyond its maximum size then cache items are deleted off the end until it returns within the memory bounds. The idea being that frequently required items are accessed regularly and won't fall off the end of the queue

### When would I want to use this?

Perhaps to only read a resource from disk once and then cache in memory, rather than repeatedly read from disk if its to be accessed multiple times

### Quick Example

```
  var (
  	Cache = memcache.CreateLRUCache(1024*4)
  	Resource = "https://raw.githubusercontent.com/seanjohnno/goexamples/master/helloworld.txt"
  )
  
  func main() {
  	fileContent, _ := GetHttpData(Resource)
  	fmt.Println("Content:", string(fileContent))
  
  	fileContent, _ = GetHttpData(Resource)
  	fmt.Println("Content:", string(fileContent))
  }
  
  type HttpContent struct {
  	Content []byte
  }
  
  func (this *HttpContent) Size() int {
  	return len(this.Content)
  }
  
  func GetHttpData(URI string) ([]byte, error) {
  	if cached, present := Cache.Get(URI); present {
  		fmt.Println("Found in cache")
  		return cached.(*HttpContent).Content, nil
  	} else {
  		fmt.Println("Not found in cache, making network request")
  
  		// No error handling here to make example shorter
  		resp, err := http.Get(URI)
  		defer resp.Body.Close()
  		body, err := ioutil.ReadAll(resp.Body)		
  		Cache.Add(URI, &HttpContent{ Content: body})
  		return body, err
  	}
  }
```

### Full Example

You can see an example with imports and stuff [here](https://github.com/seanjohnno/goexamples/blob/master/lrucache_example.go). Continue reading if you require instructions on how to grab the sourcecode and/or example from within the command-line...

### Setup

Create your Go folder structure on the filesystem (if you have't already):

```
GoProjects
  |- src
  |- pkg
  |- bin
```
In your command-line set your **GOPATH** environment variable:

* Linux: `export GOPATH=<Replace_me_with_path_to>\GoProjects`
* Windows: `set GOPATH="<Replace_me_with_path_to>\GoProjects"`

Browse to your *GoProjects* folder in the command-line and enter:

  `go get github.com/seanjohnno/memcache`

You should see the folders */github.com/seanjohnno/memcache* under *src* and the code inside *memcache*

If you want to run the example then make sure you're in your *GoProjects* folder and run:

  `go get github.com/seanjohnno/goexamples`

Navigate to the *goexamples directory* and run the following:

```
  go build lrucache_example.go
```

...and then depending on your OS:

* Linux: `./lrucache_example`
* Windows: `lrucache_example.exe`
