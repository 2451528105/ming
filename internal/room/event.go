package room

type tag string

const (
	Tag_RoomCreate tag = "room_create"
	Tag_RoomJoin   tag = "room_join"
	Tag_RoomLeave  tag = "room_leave"
	Tag_RoomClose  tag = "room_close"
	Tag_RoomReady  tag = "room_ready"
)

type Event struct {
	tag tag
	uid int64
}

func NewEvent(tag tag, uid int64) *Event {
	return &Event{
		tag: tag,
		uid: uid,
	}
}
