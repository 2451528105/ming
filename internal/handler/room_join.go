package handler

import (
	"ming/internal/game"
	"ming/internal/pb"
	"ming/sdk/consts"
	"ming/sdk/xlog"
)

// RoomJoin 进入房间。
func RoomJoin(ctx game.Context, req *pb.JoinRoomRequest) {
	if !EnsureAppGame(ctx) {
		return
	}
	xlog.Info().Int("Player", int(ctx.Uid())).Str("Route", ctx.Route()).Msgf("[Request] RoomJoin %v", req)
	if req.GetRoomId() == "" {
		ctx.ErrResp(consts.ErrorCode_RequestErr, "room_id required")
		return
	}

	ctx.OkResp(&pb.JoinRoomResponse{})
}
