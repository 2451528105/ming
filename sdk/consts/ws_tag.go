package consts

type WsTag string

const (
	WsTag_LoginGame WsTag = "login_game"
)
const (
	WsTag_Push  = 1 // 主推
	WsTag_Reply = 2 // 响应
)
