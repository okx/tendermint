package state

import "github.com/spf13/viper"

const FlagCloseStateValidation = "debug-close-state-validation"

func GetCloseStateValidation() bool {
	return viper.GetBool(FlagCloseStateValidation)
}
