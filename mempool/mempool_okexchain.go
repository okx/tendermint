package mempool

import (
	"strconv"
)

var (
	MaxTxNumPerBlock = "300"
)

//	query the MaxTxNumPerBlock from app
func (mem *CListMempool) GetMaxTxNumPerBlock() int64 {
	startBlockHeight, err := strconv.ParseInt(MaxTxNumPerBlock, 10, 64)
	if err != nil {
		return 300
	}
	return startBlockHeight
}
