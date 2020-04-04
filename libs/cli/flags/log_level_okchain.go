package flags

import "github.com/tendermint/tendermint/libs/log"

func UpdateLogLevel(lvl string) error {
	return log.UpdateLogLevel(lvl)
}
