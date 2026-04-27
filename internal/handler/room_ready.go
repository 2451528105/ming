package handler

import (
	"ming/internal/game"
	"ming/internal/pb"
	"ming/sdk/consts"
	"ming/sdk/xlog"
)

// RoomReady 准备 / 取消准备。
func RoomReady(ctx game.Context, req *pb.RoomReadyRequest) {
	if !EnsureAppGame(ctx) {
		return
	}
	xlog.Info().Int("Player", int(ctx.Uid())).Str("Route", ctx.Route()).Msgf("[Request] RoomReady %v", req)
	if req.GetRoomId() == "" {
		ctx.ErrResp(consts.ErrorCode_RequestErr, "room_id required")
		return
	}
	ctx.OkResp(&pb.RoomReadyResponse{})
}
