package room

import "errors"

var (
	ErrRoomClosed    = errors.New("room is closed")
	ErrNotInstalled  = errors.New("room default manager not installed")
	ErrInvalidRoomID = errors.New("room: empty room id")
)
