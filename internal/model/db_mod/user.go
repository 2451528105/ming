package db_mod

import "time"

type User struct {
	ID         int64     `json:"id"`
	Name       string    `json:"name"`
	Password   string    `json:"password"`
	Avatar     string    `json:"avatar"`
	Email      string    `json:"email"`
	Phone      string    `json:"phone"`
	Sex        int32     `json:"sex"`
	CreateTime time.Time `json:"create_time"`
	UpdateTime time.Time `json:"update_time"`
}
