package mempool

import (
	abci "github.com/tendermint/tendermint/abci/types"
)

const (
	MaxTxNumPerBlock = 1000
)

//	query the MaxTxNumPerBlock from app
func (mem *CListMempool) GetMaxTxNumPerBlock() int {
	res, err := mem.proxyAppQueryConn.QuerySync(
		abci.RequestQuery{
			Path: "/custom/gov/params/tendermint",
		})
	if err != nil || res.Value == nil {
		return MaxTxNumPerBlock
	}

	var num int
	cdc.MustUnmarshalJSON(res.Value, &num)
	return num
}
