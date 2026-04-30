package internal

import (
	"log"
	"ming/internal/game"
	"ming/internal/handler/ws"
	"ming/internal/room"
	"ming/sdk/consts"
)

var G *game.Game = &game.Game{}

func Init() {
	if err := room.InitDefault(); err != nil {
		log.Fatalf("init room pipeline: %v", err)
	}

	// 路由须在 G.Init() 阻塞前注册，否则 WebSocket 入站消息无法命中处理器
	G.RegisterRoute(consts.Version_V1, consts.WsTag_LoginGame, game.Handler(ws.LoginGame))
	G.RegisterRoute(consts.Version_V1, consts.WsTag_RoomCreate, game.Handler(ws.RoomCreate))
	G.RegisterRoute(consts.Version_V1, consts.WsTag_RoomJoin, game.Handler(ws.RoomJoin))
	G.RegisterRoute(consts.Version_V1, consts.WsTag_RoomExit, game.Handler(ws.RoomExit))
	G.RegisterRoute(consts.Version_V1, consts.WsTag_RoomReady, game.Handler(ws.RoomReady))
	G.Init()
}
