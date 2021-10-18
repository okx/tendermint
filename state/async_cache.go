package state

type AsyncCache struct {
	mem map[string][]byte
}

func NewAsyncCache() *AsyncCache {
	return &AsyncCache{mem: make(map[string][]byte, 0)}
}

func (a *AsyncCache) Push(key, value []byte) {
	a.mem[string(key)] = value
}

func (a *AsyncCache) Has(key []byte) bool {
	_, ok := a.mem[string(key)]
	return ok
}
