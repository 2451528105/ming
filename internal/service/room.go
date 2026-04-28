package service

import "github.com/google/uuid"

// 房间创建
func RoomCreate() (string, error) {
	return uuid.New().String(), nil
}

// 房间加入
func RoomJoin(uid int64, roomID string) {}

// 房间离开
func RoomLeave(uid int64, roomID string) {}

// 房间关闭
func RoomClose(uid int64, roomID string) {}

// 房间准备
func RoomReady(uid int64, roomID string) {}
