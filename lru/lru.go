package lru

import "container/list"

// LRU cache is not safe for concurrent access.
type Cache struct {
	maxBytes int64      // max memory can be used
	nbytes   int64      // current memory used
	ll       *list.List // double linked list
	cache    map[string]*list.Element
	// key is string, value is pointer to list.Element
	// optional and executed when an entry is purged.
	OnEvicted func(key string, value Value)
	// callback function when an entry is purged, default is nil.
}

// double linked list node
type entry struct {
	key   string
	value Value
}

// Value use Len to count how many bytes it takes.
type Value interface {
	Len() int
}

// New is the constructor of Cache.
func New(maxBytes int64, onEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

// Get look up a key's value in cache.
func (c *Cache) Get(key string) (value Value, ok bool) {
	// ele is a pointer to list.Element
	if ele, ok := c.cache[key]; ok {
		// move the element to the front of the list
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		return kv.value, true
		// if the key is in the cache, move the corresponding list.Element to the front of the list
	}
	return
}

func (c *Cache) RemoveOldest() {
	ele := c.ll.Back()
	if ele != nil {
		// remove the element from the list
		c.ll.Remove(ele)
		// get the key-value pair
		kv := ele.Value.(*entry)
		// delete the key-value pair in the map
		delete(c.cache, kv.key)
		// update the memory used
		c.nbytes -= int64(len(kv.key)) + int64(kv.value.Len())
		if c.OnEvicted != nil {
			// call the callback function
			c.OnEvicted(kv.key, kv.value)
		}
	}
}

// Add adds a value to the cache.
func (c *Cache) Add(key string, value Value) {
	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		c.nbytes += int64(value.Len()) - int64(kv.value.Len())
		kv.value = value
	} else {
		ele := c.ll.PushFront(&entry{key, value})
		c.cache[key] = ele
		c.nbytes += int64(len(key)) + int64(value.Len())
	}
	for c.maxBytes != 0 && c.maxBytes < c.nbytes {
		c.RemoveOldest()
	}
}

func (c *Cache) Len() int {
	return c.ll.Len()
}
