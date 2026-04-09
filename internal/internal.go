package internal

import (
	"ming/internal/game"
	"ming/sdk/xlog"
)

var G *game.Game = &game.Game{}

func Init() {
	//初始化游戏
	G.Init()

	xlog.Info().Msgf("游戏初始化完成")

}
