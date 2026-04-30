package ws

import (
	"ming/internal/game"
	"ming/internal/pb"
	"ming/internal/room"
	"ming/sdk/consts"
	"ming/sdk/xlog"
)

// RoomCreate 创建房间：room 内向 service 要号再登记运行时；handler 只组包回应。
func RoomCreate(ctx game.Context, req *pb.CreateRoomRequest) {
	xlog.Info().Int("Player", int(ctx.Uid())).Str("Route", ctx.Route()).Msgf("[Request] RoomCreate %v", req)

	if req.GetMaxPlayers() <= 0 {
		ctx.ErrResp(consts.ErrorCode_RequestErr, "max_players must be positive")
		return
	}

	roomID, err := room.CreateRoom()
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
