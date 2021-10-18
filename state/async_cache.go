package state

import "unsafe"

type CacheValue struct {
	value []byte
}

type AsyncCache struct {
	mem map[string]CacheValue
}

func bytes2str(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}

func NewAsyncCache() *AsyncCache {
	return &AsyncCache{mem: make(map[string]CacheValue, 0)}
}

func (a *AsyncCache) Push(key, value []byte) {
	a.mem[bytes2str(key)] = CacheValue{value: value}
}

func (a *AsyncCache) Has(key []byte) bool {
	_, ok := a.mem[bytes2str(key)]
	return ok
}
