package engine

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

func NewRedisClent(addr, clientName, username, password string, db int) (redis.UniversalClient, error) {
	rc := redis.NewUniversalClient(&redis.UniversalOptions{
		Addrs:      []string{addr},
		ClientName: clientName,
		Username:   username,
		Password:   password,
		DB:         db,
	})
	if err := rc.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("could not connect to redis: %v", err)
	}

	return rc, nil
}
