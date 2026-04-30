package ws

import (
	"ming/internal/game"
	"ming/internal/pb"
	"ming/internal/service"
	"ming/sdk/consts"
	"ming/sdk/xlog"
	"strconv"
	"time"
)

// 处理登录本游戏请求。
func LoginGame(ctx game.Context, req *pb.LoginGameRequest) {
	xlog.Info().Int("Player", int(ctx.Uid())).Str("Route", ctx.Route()).Timestamp().Msgf("[Request] LoginGame ... %v", req.String())

	userId, err := service.GetUserIdByToken(req.Token)
	if err != nil {
		ctx.ErrResp(consts.ErrorCode_RequestErr, "token is invalid")
		return
	}
	if userId != ctx.Uid() {
		ctx.ErrResp(consts.ErrorCode_RequestErr, "token is not match")
		return
	}
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
