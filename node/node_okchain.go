package node


import (
	"fmt"
	"github.com/spf13/viper"
	"github.com/tendermint/tendermint/store"
	_ "net/http/pprof"
	"path/filepath"
	"strconv"


	cfg "github.com/tendermint/tendermint/config"
	"github.com/tendermint/tendermint/types"
	dbm "github.com/tendermint/tm-db"
	sm "github.com/tendermint/tendermint/state"
	"github.com/tendermint/tendermint/libs/log"
)


func initRecover(blockStore *store.BlockStore,
	state sm.State,
	config *cfg.Config,
	blockExec *sm.BlockExecutor,
	logger log.Logger) {

	var recoverFlag bool = false
	var recoverHeight int64 = 0
	rollBack := viper.GetString(FlagRollback)
	if len(rollBack) != 0 {
		height, err := strconv.ParseInt(rollBack, 10, 64)
		if err == nil {
			recoverFlag = true
			recoverHeight = height
		}
	}

	var first, second *types.Block
	for ; recoverFlag && blockStore.Height() < recoverHeight; {
		first, second = fetchTwoBlocks(config, recoverHeight, blockStore.Height()+1)
		logger.Info("fetch blocks from recover store", "first", first.Height, "second", second.Height)

		if first == nil || second == nil {
			logger.Error("something wrong with the recover store, cannot have nil blocks")
			panic("Failed to fetch blocks from recover store due to nil block")
		}

		firstParts := first.MakePartSet(types.BlockPartSizeBytes)
		firstPartsHeader := firstParts.Header()
		firstID := types.BlockID{Hash: first.Hash(), PartsHeader: firstPartsHeader}

		err := state.Validators.VerifyCommit(
			state.ChainID, firstID, first.Height, second.LastCommit)
		if err != nil {
			logger.Error(fmt.Sprintf("fail to verify block commits at block height %v: %v", first.Height, err))
			panic(fmt.Sprintf("Failed to verify block commits at block height %v: %v", first.Height, err))
		} else {
			// TODO: batch saves so we dont persist to disk every block
			blockStore.SaveBlock(first, firstParts, second.LastCommit)

			// TODO: same thing for app - but we would need a way to
			// get the hash without persisting the state
			var err error
			state, err = blockExec.ApplyBlock(state, firstID, first)
			if err != nil {
				// TODO This is bad, are we zombie?
				panic(fmt.Sprintf("Failed to process committed block (%d:%X): %v", first.Height, first.Hash(), err))
			}
		}
	}
}


func fetchTwoBlocks(config *cfg.Config, maxHeight, height int64) (first *types.Block, second *types.Block) {
	// homePath := viper.GetString("home")
	// if homePath == "" {
	// 	homePath = "$HOME/.okchaind"
	// }
	// path := filepath.Join(os.ExpandEnv(homePath), cfg.DefaultBaseConfig().DBDir())

	path := filepath.Join(config.RootDir, "data")
	// dir, err := ioutil.ReadDir(path)
	// if err != nil {
	// 	return nil, nil
	// }
	// for _, file := range dir {
	// 	fmt.Println(file)
	// }

	// Get BlockStore
	blockStoreDB := dbm.NewDB("recoverstore", dbm.GoLevelDBBackend, path)
	defer blockStoreDB.Close()
	recoverStore := store.NewRecoverStore(blockStoreDB, maxHeight+1)

	first = recoverStore.LoadBlock(height)
	second = recoverStore.LoadBlock(height + 1)

	return
}

