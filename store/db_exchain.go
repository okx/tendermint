package store

import (
	"io/ioutil"
	"log"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/syndtr/goleveldb/leveldb/util"

	dbm "github.com/tendermint/tm-db"
)

const Interval int64 = 5

type BlockDB struct {
	db       dbm.DB
	history  map[int64]dbm.DB
	interval int64

	name    string
	backend dbm.BackendType
	dir     string
	hisDir  string

	mtx sync.Mutex
}

var _ dbm.DB = (*BlockDB)(nil)

func NewBlockDB(name string, backend dbm.BackendType, dir string) *BlockDB {
	db := dbm.NewDB(name, backend, dir)

	historyDBs := make(map[int64]dbm.DB)

	hisDir := filepath.Join(dir, "block_history")
	fs, err := ioutil.ReadDir(hisDir)
	if err != nil && !os.IsNotExist(err) {
		panic(err)
	}

	var min, max int64 = math.MaxInt64, 0
	for _, f := range fs {
		if f.IsDir() {
			index, err := strconv.ParseInt(strings.Split(f.Name(), ".")[0], 10, 64)
			if err != nil {
				continue
			}
			historyDBs[index/Interval] = dbm.NewDB(strconv.FormatInt(index, 10), backend, hisDir)
			if index < min {
				min = index
			}
			if index > max {
				max = index
			}
		}
	}

	// Check if blockdb is continuous
	if len(historyDBs) != 0 && int64(len(historyDBs)) != (max-min)/Interval+1 {
		log.Println("The block history db is discontinuous")
	}

	return &BlockDB{
		db:      db,
		history: historyDBs,

		interval: Interval,
		name:     name,
		backend:  backend,
		dir:      dir,
		hisDir:   hisDir,
	}
}

func (bdb *BlockDB) Split(height int64) {
	bdb.mtx.Lock()
	defer bdb.mtx.Unlock()

	if bdb.interval > 0 && height%bdb.interval == 0 {
		err := bdb.db.Close()
		if err != nil {
			panic(err)
		}

		if err := os.Mkdir(bdb.hisDir, 0755); err != nil && !os.IsExist(err) {
			panic(err)
		}

		hisDBName := strconv.FormatInt(height-1, 10)
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
					log.Println(err)
				}
			}
		}()
		bdb.history[(height-1)/Interval] = hisDB
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

	h := getHeightFromKey(key)
	if h > 0 {
		if db, ok := bdb.history[h/Interval]; ok {
			return db.Get(key)
		}
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

	h := getHeightFromKey(key)
	if h > 0 {
		if db, ok := bdb.history[h/Interval]; ok {
			return db.Has(key)
		}
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

	h := getHeightFromKey(key)
	if h > 0 {
		if db, ok := bdb.history[h/Interval]; ok {
			return db.Delete(key)
		}
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

	h := getHeightFromKey(key)
	if h > 0 {
		if db, ok := bdb.history[h/Interval]; ok {
			return db.DeleteSync(key)
		}
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

// getHeightFromKey
// H:{height}
// P:{height}:{partIndex}
// C:{height}
// SC:{height}
// BH:{hash}
func getHeightFromKey(key []byte) int64 {
	s := strings.Split(string(key), ":")
	if len(s) < 2 {
		return 0
	}
	h, err := strconv.ParseInt(s[1], 10, 64)
	if err != nil {
		return 0
	}
	return h
}
