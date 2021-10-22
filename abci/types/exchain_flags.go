package types

import "github.com/spf13/viper"

const (
	FlagCloseMutex   = "close-mutex"
	FlagSplitBlockDB = "split-blockdb"
)

func GetCloseMutex() bool {
	return viper.GetBool(FlagCloseMutex)
}

func SplitBlockDB() bool {
	return viper.GetBool(FlagSplitBlockDB)
}
