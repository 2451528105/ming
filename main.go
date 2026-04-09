package main

import (
	"ming/internal"
	"ming/internal/config"
	"ming/sdk"
)

func main() {
	err := sdk.ReadConfig("config/config.yaml", config.Cfg, true)
	if err != nil {
		panic(err)
	}
	internal.Init()
}
