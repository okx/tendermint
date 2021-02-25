package mempool

import (
	"strconv"
	"sync"
)

var (
	MaxTxNumPerBlockStr       = "300"
	MaxTxNum            int64 = 300
	once                sync.Once
)

func string2number(input string, defaultRes int64) int64 {
	if len(input) == 0 {
		return defaultRes
	}

	res, err := strconv.ParseInt(input, 10, 64)
	if err != nil {
		panic(err)
	}
	return res
}

func init() {
	once.Do(func() {
		MaxTxNum = string2number(MaxTxNumPerBlockStr, 300)
	})
}

//	query the MaxTxNumPerBlock from app
func (mem *CListMempool) GetMaxTxNumPerBlock() int64 {
	return MaxTxNum
}
