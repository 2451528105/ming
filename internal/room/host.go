package room

import (
	"errors"

	"github.com/google/uuid"
)

// 进程内默认 Pipeline，在组装根调用 InitDefault 后可用；供 handler 等在不依赖 game 的前提下使用 room 能力。

var hosted *Pipeline

// ErrHostedNotReady 表示未先调用 InitDefault。
var ErrHostedNotReady = errors.New("room: default pipeline not initialized")

// InitDefault 使用 NewDefaultPipeline 初始化并挂载默认管道（整个进程调用一次）。
func InitDefault() error {
	p, err := NewDefaultPipeline()
	if err != nil {
		return err
	}
	hosted = p
	return nil
}

// HostedNewRoom 生成新的房间 ID、在默认管道上创建对应 Actor，并返回该 ID（由 room 层负责 ID，不经过 Game）。
func HostedNewRoom() (roomID string, err error) {
	if hosted == nil {
		return "", ErrHostedNotReady
	}
	id := uuid.New().String()
	hosted.CreateRoom(id)
	return id, nil
}

// HostedCreateRoom 在默认管道上为已有 roomID 创建/获取串行 Actor（加入房间等场景可继续用）。
func HostedCreateRoom(roomID string) {
	if hosted != nil {
		hosted.CreateRoom(roomID)
	}
}

// CloseDefault 关闭默认管道（通常在进程退出钩子中调用）。
func CloseDefault() {
	if hosted != nil {
		hosted.Close()
		hosted = nil
	}
}
