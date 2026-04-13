package locate

import (
	"context"
	"fmt"
	"strconv"

	"github.com/redis/go-redis/v9"
)

var unbindGateLua = redis.NewScript(`
local key = KEYS[1]
local nodeField = ARGV[1]
local connField = ARGV[2]
local expectNode = ARGV[3]
local expectConn = ARGV[4]

local vals = redis.call('HMGET', key, nodeField, connField)
local curNode = vals[1]
local curConn = vals[2]

if curNode == expectNode and curConn == expectConn then
	redis.call('HMSET', key, nodeField, '', connField, '')
	return 1
end

return 0
`)

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

func (r *RedisLocator) UnbindGateNode(uid int64, node string, connId int64) error {
	key := fmt.Sprintf(r.playerKeyFormat, uid)
	expectConnId := strconv.FormatInt(connId, 10)
	_, err := unbindGateLua.Run(
		context.Background(),
		r.gameRedis,
		[]string{key},
		r.gateNodeField,
		r.gateConnIdField,
		node,
		expectConnId,
	).Int()
	return err
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
