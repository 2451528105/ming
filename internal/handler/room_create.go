package handler

import (
	"ming/internal/game"
	"ming/internal/pb"
	"ming/internal/room"
	"ming/sdk/consts"
	"ming/sdk/xlog"
)

// RoomCreate 创建房间：在 room 层注册房间 Actor，并返回房间信息。
func RoomCreate(ctx game.Context, req *pb.CreateRoomRequest) {
	if !EnsureAppGame(ctx) {
		return
	}
	xlog.Info().Int("Player", int(ctx.Uid())).Str("Route", ctx.Route()).Msgf("[Request] RoomCreate %v", req)

	if req.GetMaxPlayers() <= 0 {
		ctx.ErrResp(consts.ErrorCode_RequestErr, "max_players must be positive")
		return
	}

	roomID, err := room.HostedNewRoom()
	if err != nil {
		ctx.ErrResp(consts.ErrorCode_RequestErr, err.Error())
		return
	}

	ctx.OkResp(&pb.CreateRoomResponse{
		RoomInfo: &pb.RoomInfo{
			RoomId:     roomID,
			Name:       req.GetName(),
			MaxPlayers: req.GetMaxPlayers(),
			CurPlayers: 1,
			OwnerUid:   ctx.Uid(),
			Status:     0,
		},
	})
}
