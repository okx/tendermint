package store

import (
	"fmt"
	"io/ioutil"
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
	hisDir  string

	mtx sync.Mutex
}

var _ dbm.DB = (*BlockDB)(nil)

func NewBlockDB(name string, backend dbm.BackendType, dir string) *BlockDB {
	db := dbm.NewDB(name, backend, dir)

	var history []dbm.DB

	hisDir := filepath.Join(dir, "block_history")
	fs, err := ioutil.ReadDir(hisDir)
	if err != nil && !os.IsNotExist(err) {
		panic(err)
	}
	for _, f := range fs {
		if f.IsDir() {
			fmt.Println(f.Name())
			history = append(history, dbm.NewDB(f.Name(), backend, hisDir))
		}
	}

	return &BlockDB{
		db:      db,
		history: history,

		interval: 5,
		name:     name,
		backend:  backend,
		dir:      dir,
		hisDir:   hisDir,
	}
}

func (bdb *BlockDB) Split(height int64) {
	bdb.mtx.Lock()
	defer bdb.mtx.Unlock()

	if bdb.interval > 0 && height%int64(bdb.interval) == 0 {
		err := bdb.db.Close()
		if err != nil {
			panic(err)
		}

		if err := os.Mkdir(bdb.hisDir, 0755); err != nil && !os.IsExist(err) {
			panic(err)
		}

		hisDBName := strconv.FormatInt(height, 10)
		hisPath := filepath.Join(bdb.hisDir, hisDBName+".db")
		dbPath := filepath.Join(bdb.dir, bdb.name+".db")
		err = os.Rename(dbPath, hisPath)
		if err != nil {
			panic(err)
		}

		hisDB := dbm.NewDB(hisDBName, bdb.backend, bdb.hisDir)
		go func() {
			if ldb, ok := hisDB.(*dbm.GoLevelDB); ok {
				err = ldb.DB().CompactRange(util.Range{})
				if err != nil {
					fmt.Println(err)
				}
			}
		}()
		bdb.history = append(bdb.history, hisDB)
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
