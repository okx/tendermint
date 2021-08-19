package state

type CacheValue struct {
	value []byte
}

type AsyncCache struct {
	mem map[string]CacheValue
}

func NewAsyncCache() *AsyncCache {
	return &AsyncCache{mem: make(map[string]CacheValue, 0)}
}

func (a *AsyncCache) Push(key, value []byte) {
	a.mem[string(key)] = CacheValue{value: value}
}

func (a *AsyncCache) Has(key []byte) bool {
	_, ok := a.mem[string(key)]
	return ok
}
