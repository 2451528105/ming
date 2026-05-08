package manager

import (
	"context"
	"errors"
	"fmt"
	"ming/internal/model/db_mod"
	"ming/sdk/consts"
	"time"

	"gorm.io/gorm"
)

// 根据id获取用户
func GetUserById(id int64) (*db_mod.User, error) {
	if m == nil {
		return nil, errors.New("manager is not initialized")
	}
	db := m.db
	if db == nil {
		return nil, errors.New("gorm db is not initialized")
	}
	user := &db_mod.User{}
	err := db.Table(`"User"`).Where("id = ?", id).First(user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return user, nil
}

// 创建用户
func CreateUser(user *db_mod.User) error {
	if m == nil {
		return errors.New("manager is not initialized")
	}
	db := m.db
	if db == nil {
		return errors.New("gorm db is not initialized")
	}
	return db.Table(`"User"`).Create(user).Error
}

// 保存用户token
func SaveUserToken(id int64, token string) error {
	if m == nil {
		return errors.New("manager is not initialized")
	}
	redis := m.redis
	if redis == nil {
		return errors.New("redis is not initialized")
	}
	return redis.Set(context.Background(), fmt.Sprintf(consts.KeyFormat_AuthToken, token), id, time.Minute*30).Err()
}

// 根据token获取用户id
func GetUserIdByToken(token string) (int64, error) {
	if m == nil {
		return 0, errors.New("manager is not initialized")
	}
	redis := m.redis
	if redis == nil {
		return 0, errors.New("redis is not initialized")
	}
	return redis.Get(context.Background(), fmt.Sprintf(consts.KeyFormat_AuthToken, token)).Int64()
}
