package config


type IDynamicConfig interface {
	GetTpb() int64
}

var DynamicConfig IDynamicConfig

func SetConfig(c IDynamicConfig) {
	DynamicConfig = c
}
