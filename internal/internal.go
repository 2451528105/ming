package internal

import (
	"ming/internal/game"
)

var G *game.Game = &game.Game{}

func Init() {
	//初始化游戏
	G.Init()

}
