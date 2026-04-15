package game

import (
	"database/sql"
	"fmt"
	"ming/internal/config"
	"ming/internal/room"
	"ming/internal/service"
	"ming/sdk/consts"
	"ming/sdk/engine"
	"ming/sdk/locate"
	"ming/sdk/netutil"
	"ming/sdk/rand"
	"ming/sdk/snowflake"
	"ming/sdk/transceiver"
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
	nodeId      string                  // 节点id
	ip          string                  // ip地址
	redis       redis.UniversalClient   // redis客户端
	postgres    *sql.DB                 // postgres数据库客户端
	producer    rmq.Producer            // rmq生产者
	consumer    rmq.SimpleConsumer      // rmq消费者
	wsServer    *melody.Melody          // websocket服务器
	idGen       *snowflake.Generator    // 雪花算法生成器
	transceiver transceiver.Transceiver // rmq消息接收器
	locator     locate.Locator          // 节点定位器
	sessions    sync.Map                // 用户会话存储器：key: userId(int64), value: *melody.Session

	urm      *UserRequestManager // 用户请求管理器
	roomPipe *room.Pipeline
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
	g.transceiver = transceiver.NewXRMQTransceiver(consts.TopicGameMessage, g.producer, g.consumer)
	g.locator = locate.NewLocator(g.redis, consts.KeyFormat_Player, consts.Field_GateNode, consts.Field_GateConnId)
	g.roomPipe, err = room.Init(room.PipelineOptions{
		MailboxSize: 256,
		Resolver:    room.NewUIDResolver(),
		Service:     service.NewDefaultRoomService(),
	})
	if err != nil {
		xlog.Error().Err(err).Msg("[Init] init room pipeline failed")
		return
	}
	g.urm = NewUserRequestManager(func(event *requestEvent) {
		header := event.data.GetHeader()
		if header == nil {
			xlog.Error().Msg("[Init] room dispatch missing header")
			return
		}
		if err := g.roomPipe.Dispatch(header.GetUid(), event.data.GetRoute(), event.data.GetPayload(), header.GetMsgId()); err != nil {
			xlog.Error().Err(err).Msgf("[Init] room dispatch failed, uid: %d", header.GetUid())
		}
	})
	//启动各个组件
	g.startConnectionServices()
}

// 启动各个组件
func (g *Game) startConnectionServices() error {
	g.wsServer = melody.New()
	//1.监听游戏消息
	g.listenGameMessage()
	//2.配置webSocket服务
	g.configureWebSocket()
	//3.配置http服务
	//4.启动http服务器
	return nil
}

// 监听游戏消息 (其它节点发送给本节点玩家的消息)
func (g *Game) listenGameMessage() {
	err := g.transceiver.ReceiveMessage(config.Cfg.Application.Name, g.nodeId, func(uid int64, payload []byte, msgId string) error {
		xlog.Info().Msgf("[listenGameMessage] ReceiveMessage, uid: %d; msgId: %s", uid, msgId)
		s, ok := g.sessions.Load(uid)
		if !ok {
			return fmt.Errorf("session not found, uid: %d", uid)
		}
		return s.(*melody.Session).WriteBinary(payload)
	})
	if err != nil {
		xlog.Error().Msgf("[listenGameMessage] ReceiveMessage error: %v", err)
	}
}
