package store

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	"github.com/syndtr/goleveldb/leveldb/util"

	dbm "github.com/tendermint/tm-db"
)

type BlockDB struct {
	db       dbm.DB
	history  []dbm.DB
	interval uint64

	name    string
	backend dbm.BackendType
	dir     string

	mtx sync.Mutex
}

var _ dbm.DB = (*BlockDB)(nil)

func NewBlockDB(db dbm.DB) *BlockDB {
	return &BlockDB{
		db:       db,
		interval: 5,
		name:     "blockstore",
		backend:  dbm.GoLevelDBBackend,
		dir:      "tools/_cache_evm/data",
	}
}

func (bdb *BlockDB) Split(height int64) {
	bdb.mtx.Lock()
	defer bdb.mtx.Unlock()

	if bdb.interval > 0 && height%int64(bdb.interval) == 0 {
		err := bdb.db.Close()
		if err != nil {
			fmt.Println(err)
		}
		err = os.Rename(filepath.Join(bdb.dir, bdb.name+".db"), filepath.Join(bdb.dir, strconv.FormatInt(height, 10)+".db"))
		if err != nil {
			fmt.Println(err)
		}

		oldDB := dbm.NewDB(strconv.FormatInt(height, 10), bdb.backend, bdb.dir)
		go func() {
			if ldb, ok := oldDB.(*dbm.GoLevelDB); ok {
				err = ldb.DB().CompactRange(util.Range{})
				if err != nil {
					fmt.Println(err)
				}
			}
		}()
		bdb.history = append(bdb.history, oldDB)
		bdb.db = dbm.NewDB(bdb.name, bdb.backend, bdb.dir)
	}
}

// Get implements DB.
func (bdb *BlockDB) Get(key []byte) ([]byte, error) {
	bdb.mtx.Lock()
	defer bdb.mtx.Unlock()

	b, err := bdb.db.Get(key)
	if b != nil || err != nil {
		return b, err
	}
	for _, db := range bdb.history {
		b, err = db.Get(key)
		if b != nil || err != nil {
			return b, err
		}
	}
	return b, err
}

// Has implements DB.
func (bdb *BlockDB) Has(key []byte) (bool, error) {
	bdb.mtx.Lock()
	defer bdb.mtx.Unlock()

	has, err := bdb.db.Has(key)
	if has || err != nil {
		return has, err
	}
	for _, db := range bdb.history {
		has, err = db.Has(key)
		if has || err != nil {
			return has, err
		}
	}
	return has, err
}

// Set implements DB.
func (bdb *BlockDB) Set(key []byte, value []byte) error {
	bdb.mtx.Lock()
	defer bdb.mtx.Unlock()

	return bdb.db.Set(key, value)
}

// SetSync implements DB.
func (bdb *BlockDB) SetSync(key []byte, value []byte) error {
	bdb.mtx.Lock()
	defer bdb.mtx.Unlock()

	return bdb.db.SetSync(key, value)
}

// Delete implements DB.
func (bdb *BlockDB) Delete(key []byte) error {
	bdb.mtx.Lock()
	defer bdb.mtx.Unlock()

	err := bdb.db.Delete(key)
	if err != nil {
		return err
	}
	for _, db := range bdb.history {
		err = db.Delete(key)
		if err != nil {
			return err
		}
	}
	return nil
}

// DeleteSync implements DB.
func (bdb *BlockDB) DeleteSync(key []byte) error {
	bdb.mtx.Lock()
	defer bdb.mtx.Unlock()

	err := bdb.db.DeleteSync(key)
	if err != nil {
		return err
	}
	for _, db := range bdb.history {
		err = db.DeleteSync(key)
		if err != nil {
			return err
		}
	}
	return nil
}

// Close implements DB.
func (bdb *BlockDB) Close() error {
	bdb.mtx.Lock()
	defer bdb.mtx.Unlock()

	err := bdb.db.Close()
	if err != nil {
		return err
	}
	for _, db := range bdb.history {
		err = db.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

// Print implements DB.
// Not used in the BlockStore
func (bdb *BlockDB) Print() error {
	return nil
}

// Stats implements DB.
// Not used in the BlockStore
func (bdb *BlockDB) Stats() map[string]string {
	return nil
}

// NewBatch implements DB.
func (bdb *BlockDB) NewBatch() dbm.Batch {
	bdb.mtx.Lock()
	defer bdb.mtx.Unlock()

	return bdb.db.NewBatch()
}

// Iterator implements DB.
// Not used in the BlockStore
func (bdb *BlockDB) Iterator(start, end []byte) (dbm.Iterator, error) {
	return nil, nil
}

// ReverseIterator implements DB.
// Not used in the BlockStore
func (bdb *BlockDB) ReverseIterator(start, end []byte) (dbm.Iterator, error) {
	return nil, nil
}
