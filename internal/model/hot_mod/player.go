package hot_mod

type Player struct {
	ID      int64  `json:"id"`
	Name    string `json:"name"`
	Avatar  string `json:"avatar"`
	Offline int32  `json:"offline"` // 0: 在线 1: 离线
}
