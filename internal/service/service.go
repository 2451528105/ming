package service

import "ming/sdk/xlog"

// RoomService is the business layer called by room actor.
// Keep it stateless where possible.
type RoomService interface {
	Handle(uid int64, route string, payload []byte, msgID string)
}

type DefaultRoomService struct{}

func NewDefaultRoomService() *DefaultRoomService {
	return &DefaultRoomService{}
}

func (s *DefaultRoomService) Handle(uid int64, route string, payload []byte, msgID string) {
	xlog.Info().Msgf("[service] handle room event, uid: %d, route: %s, msgId: %s, payloadLen: %d", uid, route, msgID, len(payload))
}
