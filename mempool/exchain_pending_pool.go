package mempool

import (
	"fmt"
	"sync"
)

type PendingPool struct {
	maxSize       int
	addressTxsMap map[string]map[uint64]*PendingTx
	txsMap        map[string]*PendingTx
	mtx           sync.RWMutex
	period        int
	periodLimit   int
	periodCounter map[string]int // address with period count
}

func newPendingPool(maxSize int, period int, periodLimit int) *PendingPool {
	return &PendingPool{
		maxSize:       maxSize,
		addressTxsMap: make(map[string]map[uint64]*PendingTx),
		txsMap:        make(map[string]*PendingTx),
		period:        period,
		periodLimit:   periodLimit,
		periodCounter: make(map[string]int),
	}
}

func (p *PendingPool) Size() int {
	p.mtx.RLock()
	defer p.mtx.RUnlock()
	return len(p.txsMap)
}

func (p *PendingPool) getTx(address string, nonce uint64) *PendingTx {
	p.mtx.RLock()
	defer p.mtx.RUnlock()
	if _, ok := p.addressTxsMap[address]; ok {
		return p.addressTxsMap[address][nonce]
	}
	return nil
}

func (p *PendingPool) addTx(pendingTx *PendingTx) {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	if _, ok := p.addressTxsMap[pendingTx.exTxInfo.Sender]; !ok {
		p.addressTxsMap[pendingTx.exTxInfo.Sender] = make(map[uint64]*PendingTx)
	}
	p.addressTxsMap[pendingTx.exTxInfo.Sender][pendingTx.exTxInfo.Nonce] = pendingTx
	p.txsMap[txID(pendingTx.mempoolTx.tx)] = pendingTx
}

func (p *PendingPool) removeTx(address string, nonce uint64) {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	if _, ok := p.addressTxsMap[address]; ok {
		if pendingTx, ok := p.addressTxsMap[address][nonce]; ok {
			delete(p.addressTxsMap[address], nonce)
			delete(p.txsMap, txID(pendingTx.mempoolTx.tx))
		}
		if len(p.addressTxsMap[address]) == 0 {
			delete(p.addressTxsMap, address)
			delete(p.periodCounter, address)
		}
		// update period counter
		if count, ok := p.periodCounter[address]; ok && count > 0 {
			p.periodCounter[address] = count - 1
		}

	}

}

func (p *PendingPool) removeTxByHash(txHash string) {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	if pendingTx, ok := p.txsMap[txHash]; ok {
		delete(p.txsMap, txHash)
		if _, ok := p.addressTxsMap[pendingTx.exTxInfo.Sender]; ok {
			delete(p.addressTxsMap[pendingTx.exTxInfo.Sender], pendingTx.exTxInfo.Nonce)
			if len(p.addressTxsMap[pendingTx.exTxInfo.Sender]) == 0 {
				delete(p.addressTxsMap, pendingTx.exTxInfo.Sender)
				delete(p.periodCounter, pendingTx.exTxInfo.Sender)
			}
			// update period counter
			if count, ok := p.periodCounter[pendingTx.exTxInfo.Sender]; ok && count > 0 {
				p.periodCounter[pendingTx.exTxInfo.Sender] = count - 1
			}
		}
	}
}

func (p *PendingPool) handlePendingTxNonce(accRetriever AccountRetriever) map[string]uint64 {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	addrMap := make(map[string]uint64)
	for addr, txMap := range p.addressTxsMap {
		accountNonce := accRetriever.GetAccountNonce(addr)
		for nonce, pendingTx := range txMap {
			// remove invalid pending tx
			if nonce < accountNonce {
				delete(p.addressTxsMap[addr], nonce)
				delete(p.txsMap, txID(pendingTx.mempoolTx.tx))
			} else if nonce == accountNonce {
				addrMap[addr] = nonce
			}
		}
		if len(p.addressTxsMap[addr]) == 0 {
			delete(p.addressTxsMap, addr)
		}
	}
	return addrMap
}

func (p *PendingPool) handlePeriodCounter() {
	p.mtx.Lock()
	defer p.mtx.Unlock()
	for addr, txMap := range p.addressTxsMap {
		count := p.periodCounter[addr]
		if count >= p.periodLimit {
			delete(p.addressTxsMap, addr)
			for _, pendingTx := range txMap {
				delete(p.txsMap, txID(pendingTx.mempoolTx.tx))
			}
			delete(p.periodCounter, addr)
		} else {
			p.periodCounter[addr] = count + 1
		}
	}
}

func (p *PendingPool) isFull() error {
	poolSize := p.Size()
	if poolSize >= p.maxSize {
		return ErrPendingPoolIsFull{
			size:    poolSize,
			maxSize: p.maxSize,
		}
	}
	return nil
}

type PendingTx struct {
	mempoolTx *mempoolTx
	exTxInfo  ExTxInfo
}

// ErrPendingPoolIsFull means PendingPool can't handle that much load
type ErrPendingPoolIsFull struct {
	size    int
	maxSize int
}

func (e ErrPendingPoolIsFull) Error() string {
	return fmt.Sprintf(
		"PendingPool is full: current pending pool size %d, max size %d",
		e.size, e.maxSize)
}

type AccountRetriever interface {
	GetAccountNonce(address string) uint64
}
