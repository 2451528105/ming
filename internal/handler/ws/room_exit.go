package ws

import (
	"ming/internal/game"
	"ming/internal/pb"
	"ming/internal/room"
	"ming/sdk/consts"
	"ming/sdk/xlog"
)

// RoomExit 退出房间。
func RoomExit(ctx game.Context, req *pb.ExitRoomRequest) {
	xlog.Info().Int("Player", int(ctx.Uid())).Str("Route", ctx.Route()).Msgf("[Request] RoomExit %v", req)
	if req.GetRoomId() == "" {
		ctx.ErrResp(consts.ErrorCode_RequestErr, "room_id required")
		return
	}
	room.Submit(req.GetRoomId(), room.NewEvent(room.Tag_RoomLeave, ctx.Uid()))
}
