package locate

import (
	"context"
	"fmt"
	"strconv"

	"github.com/redis/go-redis/v9"
)

type RedisLocator struct {
	gameRedis       redis.UniversalClient // 游戏redis
	playerKeyFormat string                // 玩家key format
	gateNodeField   string                // 玩家gate node字段名
	gateConnIdField string                // 玩家gate connId字段名
}

func NewLocator(redisClient redis.UniversalClient, playerKeyFormat, gateNodeField, gateConnIdField string) *RedisLocator {
	return &RedisLocator{
		gameRedis:       redisClient,
		playerKeyFormat: playerKeyFormat,
		gateNodeField:   gateNodeField,
		gateConnIdField: gateConnIdField,
	}
}
func (r *RedisLocator) Name() string {
	return "redis"
}

func (r *RedisLocator) BindGateNode(uid int64, node string, connId int64) error {
	key := fmt.Sprintf(r.playerKeyFormat, uid)
	return r.gameRedis.HMSet(context.Background(), key, r.gateNodeField, node, r.gateConnIdField, connId).Err()
}

func (r *RedisLocator) UnBindGateNode(uid int64, node string) error {
	current, err := r.GetGateNode(uid)
	if err != nil {
		return err
	}
	if current == node {
		key := fmt.Sprintf(r.playerKeyFormat, uid)
		if err = r.gameRedis.HMSet(context.Background(), key, r.gateNodeField, "").Err(); err != nil {
			return err
		}
	}
	return nil
}

func (r *RedisLocator) GetGateNode(uid int64) (string, error) {
	key := fmt.Sprintf(r.playerKeyFormat, uid)
	node, err := r.gameRedis.HGet(context.Background(), key, r.gateNodeField).Result()
	if err != nil {
		return "", err
	}
	return node, nil
}
func (r *RedisLocator) GetGateConnId(uid int64) (int64, error) {
	key := fmt.Sprintf(r.playerKeyFormat, uid)
	connId, err := r.gameRedis.HGet(context.Background(), key, r.gateConnIdField).Result()
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(connId, 10, 64)
}
