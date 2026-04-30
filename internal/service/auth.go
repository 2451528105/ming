package service

import (
	"errors"
	"ming/internal/manager"
	"ming/internal/model/db_mod"
	"ming/sdk/tools"
)

func Login(id int64, password string) (string, error) {
	user, err := manager.GetUserById(id)
	if err != nil {
		return "", err
	}
	if user.Password != password {
		return "", errors.New("password is incorrect")
	}
	token := tools.GenerateToken()
	if err := manager.SaveUserToken(id, token); err != nil {
		return "", err
	}
	return token, nil
}

func Register(name, password string, sex int) (int64, error) {
	user := &db_mod.User{
		Name:     name,
		Password: password,
		Sex:      int32(sex),
	}
	if err := manager.CreateUser(user); err != nil {
		return 0, err
	}
	return user.ID, nil
}
