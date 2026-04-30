package hot_mod

type Seat struct {
	ID       int64 `json:"id"`
	PlayerID int64 `json:"player_id"`
	State    int32 `json:"state"` // 0: 空闲 1: 准备 2: 游戏 3: 离开
}
