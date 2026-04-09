package game

import (
	"database/sql"
	"ming/internal/config"
	"ming/sdk/engine"
	"ming/sdk/netutil"
	"ming/sdk/rand"
	"ming/sdk/snowflake"
	"ming/sdk/xlog"
	"sync"

	rmq "github.com/apache/rocketmq-clients/golang"
	"github.com/olahol/melody"
	"github.com/redis/go-redis/v9"
)

const (
	UserKey   = "uid"
	ConnIDKey = "conn_id"
)

type Game struct {
	nodeId   string
	ip       string
	redis    redis.UniversalClient
	postgres *sql.DB
	producer rmq.Producer
	consumer rmq.SimpleConsumer
	wsServer *melody.Melody
	idGen    *snowflake.Generator

	sessions sync.Map // key: userId(int64), value: *melody.Session
}

func (g *Game) Init() {
	//1.初始化nodeId和ip
	g.nodeId = rand.RandLittleLetter(6)
	ip, err := netutil.OutboundIPv4()
	if err != nil {
		xlog.Error().Err(err).Msg("获取ip失败")
		return
	}
	g.ip = ip
	g.idGen = snowflake.NewGeneratorByNode(g.nodeId)

	//2.初始化日志
	xlog.Init(config.Cfg.Log.MinLevel,
		config.Cfg.Log.Pathname,
		config.Cfg.Log.Interval,
		config.Cfg.Application.Name,
		config.Cfg.Application.Env,
		g.nodeId,
		g.ip,
		config.Cfg.Log.TimeFormat)

	//3.初始化redis
	redis, err := engine.NewRedisClent(config.Cfg.Redis.Addr,
		config.Cfg.Redis.ClientName,
		config.Cfg.Redis.Username,
		config.Cfg.Redis.Password,
		config.Cfg.Redis.DB)
	if err != nil {
		xlog.Error().Err(err).Msg("初始化redis失败")
		return
	}
	g.redis = redis

	//4.初始化pgsql
	postgres, err := engine.NewPostgresClient(
		config.Cfg.PostgresSql.Host,
		config.Cfg.PostgresSql.Port,
		config.Cfg.PostgresSql.User,
		config.Cfg.PostgresSql.Password,
		config.Cfg.PostgresSql.DB,
	)
	if err != nil {
		xlog.Error().Err(err).Msg("初始化pgsql失败")
		return
	}
	g.postgres = postgres

	//5.初始化rmq
	producer, consumer, err := engine.InitRmq(config.Cfg.RocketMQ.Endpoint, config.Cfg.RocketMQ.Namespace, config.Cfg.RocketMQ.ConsumerGroup)
	if err != nil {
		xlog.Error().Err(err).Msg("初始化rmq失败")
		return
	}
	g.producer = producer
	g.consumer = consumer

	//启动各个组件
	g.startConnectionServices()
}

// 启动各个组件
func (g *Game) startConnectionServices() error {
	g.wsServer = melody.New()
	//1.监听游戏消息
	//2.配置webSocket服务
	g.configureWebSocket()
	//3.配置http服务
	//4.启动http服务器
	return nil
}
