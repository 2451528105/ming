package ws

import (
	"ming/internal/game"
	"ming/internal/pb"
	"ming/sdk/xlog"
	"strconv"
	"time"
)

// 处理登录本游戏请求。
func LoginGame(ctx game.Context, req *pb.LoginGameRequest) {
	xlog.Info().Int("Player", int(ctx.Uid())).Str("Route", ctx.Route()).Timestamp().Msgf("[Request] LoginGame ... %v", req.String())

	ctx.OkResp(&pb.LoginGameResponse{
		PlayerInfo: &pb.PlayerInfo{
			Uid:          ctx.Uid(),
			Nickname:     "test",
			Avatar:       "test",
			Sex:          1,
			RegisterTime: strconv.FormatInt(time.Now().UnixMilli(), 10),
		},
	})
}
