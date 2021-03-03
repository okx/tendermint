package state

import (
	"github.com/spf13/viper"
	dbm "github.com/tendermint/tm-db"
)

var (
	PrefixStartHeight = []byte("startHeight")
)

// LoadStateStartHeight loads the State from the database.
func LoadStateStartHeight(db dbm.DB) State {
	startHeight := viper.GetString("start_height")
	if startHeight == "0" {
		return loadState(db, stateKey)
	}
	state := loadState(db, append(PrefixStartHeight, []byte(startHeight)...))
	if state.IsEmpty() {
		return loadState(db, stateKey)
	}
	return state
}

// SaveStateStartHeight persists the State, the ValidatorsInfo, and the ConsensusParamsInfo to the database.
// This flushes the writes (e.g. calls SetSync).
func SaveStateStartHeight(db dbm.DB, state State) {
	saveState(db, state, stateKey)
}
