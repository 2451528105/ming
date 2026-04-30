package hot_mod

type Room struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	Seats  []Seat `json:"seats"`
	Status int32  `json:"status"` // 0: 空闲 1: 游戏中 2: 结算中
}
