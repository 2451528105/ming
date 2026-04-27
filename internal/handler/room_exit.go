package handler

import (
	"ming/internal/game"
	"ming/internal/pb"
	"ming/sdk/consts"
	"ming/sdk/xlog"
)

// RoomExit 退出房间。
func RoomExit(ctx game.Context, req *pb.ExitRoomRequest) {
	if !EnsureAppGame(ctx) {
		return
	}
	xlog.Info().Int("Player", int(ctx.Uid())).Str("Route", ctx.Route()).Msgf("[Request] RoomExit %v", req)
	if req.GetRoomId() == "" {
		ctx.ErrResp(consts.ErrorCode_RequestErr, "room_id required")
		return
	}
	ctx.OkResp(&pb.ExitRoomResponse{RoomId: req.GetRoomId()})
}
