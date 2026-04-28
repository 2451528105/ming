package room

import (
	"sync"

	"ming/internal/service"
)

// Room 单房间一条协程，从邮箱取信号后串行处理。
type Room struct {
	id       string
	mailbox  chan *Event
	shutdown chan struct{}
	done     chan struct{}

	once sync.Once
}

func NewRoom(id string, mailboxSize int) *Room {
	if mailboxSize <= 0 {
		mailboxSize = 128
	}
	r := &Room{
		id:       id,
		mailbox:  make(chan *Event, mailboxSize),
		shutdown: make(chan struct{}),
		done:     make(chan struct{}),
	}
	go r.loop()
	return r
}

// Submit 向本房间邮箱送入一次处理事件。缓冲区满时会阻塞，直到 consumer 腾出空间或房间关闭。
func (r *Room) Submit(e *Event) error {
	// 已关闭时不再入队（优先于阻塞发送，避免关闭后仍把事件送进缓冲区）。
	select {
	case <-r.shutdown:
		return ErrRoomClosed
	default:
	}
	select {
	case r.mailbox <- e:
		return nil
	case <-r.shutdown:
		return ErrRoomClosed
	}
}

func (r *Room) Close() {
	r.once.Do(func() {
		close(r.shutdown)
		<-r.done
		close(r.mailbox)
	})
}

func (r *Room) loop() {
	defer close(r.done)
	for {
		select {
		case e, ok := <-r.mailbox:
			if !ok {
				return
			}
			r.handle(e)
		case <-r.shutdown:
			r.drainMailbox()
			return
		}
	}
}

func (r *Room) drainMailbox() {
	for {
		select {
		case e, ok := <-r.mailbox:
			if !ok {
				return
			}
			r.handle(e)
		default:
			return
		}
	}
}

func (r *Room) handle(e *Event) {
	switch e.tag {
	case Tag_RoomJoin:
		service.RoomJoin(e.uid, r.id)
	case Tag_RoomLeave:
		service.RoomLeave(e.uid, r.id)
	case Tag_RoomClose:
		service.RoomClose(e.uid, r.id)
	case Tag_RoomReady:
		service.RoomReady(e.uid, r.id)
	}
}
