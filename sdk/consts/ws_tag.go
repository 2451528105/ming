package consts

const (
	WsTag_LoginGame = "login_game"

	// 房间相关（与客户端 route 一致，用于 URM 路由 key：version:route）
	WsTag_RoomCreate = "room_create"
	WsTag_RoomJoin   = "room_join"
	WsTag_RoomExit   = "room_exit"
	WsTag_RoomReady  = "room_ready"
)
const (
	WsTag_Push  = 1 // 主推
	WsTag_Reply = 2 // 响应
)
