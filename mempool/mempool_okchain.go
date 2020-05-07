package mempool

const (
	MaxTxNumPerBlock = 2000
)

//	query the MaxTxNumPerBlock from app
func (mem *CListMempool) GetMaxTxNumPerBlock() int {
	return MaxTxNumPerBlock
}
