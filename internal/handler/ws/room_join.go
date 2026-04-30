package ws

import (
	"ming/internal/game"
	"ming/internal/pb"
	"ming/internal/room"
	"ming/sdk/consts"
	"ming/sdk/xlog"
)

// RoomJoin 进入房间。
func RoomJoin(ctx game.Context, req *pb.JoinRoomRequest) {
	xlog.Info().Int("Player", int(ctx.Uid())).Str("Route", ctx.Route()).Msgf("[Request] RoomJoin %v", req)
	if req.GetRoomId() == "" {
		ctx.ErrResp(consts.ErrorCode_RequestErr, "room_id required")
		return
	}

	room.Submit(req.GetRoomId(), room.NewEvent(room.Tag_RoomJoin, ctx.Uid()))
}
