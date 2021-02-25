package mempool

const (
	MaxTxNumPerBlock = 300
)

//	query the MaxTxNumPerBlock from app
func (mem *CListMempool) GetMaxTxNumPerBlock() int {
	return MaxTxNumPerBlock
}
