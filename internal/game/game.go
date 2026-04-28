package game

import (
	"context"
	"database/sql"
	"fmt"
	"ming/internal/config"
	"ming/sdk/consts"
	"ming/sdk/engine"
	"ming/sdk/locate"
	"ming/sdk/netutil"
	"ming/sdk/rand"
	"ming/sdk/snowflake"
	"ming/sdk/transceiver"
	"ming/sdk/xlog"
	"net/http"
	"os"
	"sync"
	"time"

	rmq "github.com/apache/rocketmq-clients/golang"
	"github.com/ivy-mobile/odin/xutil/xos"
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
	httpServer  *http.Server            // http服务器
	idGen       *snowflake.Generator    // 雪花算法生成器
	transceiver transceiver.Transceiver // rmq消息接收器
	locator     locate.Locator          // 节点定位器
	sessions    sync.Map                // 用户会话存储器：key: userId(int64), value: *melody.Session

	urm           *UserRequestManager // 用户请求管理器
	routes        sync.Map            // 路由存储器：key: 版本_tag, value: 处理逻辑
	shutdownHooks []func() // 进程退出时逆序调用，由组装层注册
}

// RegisterShutdownHook 注册进程关闭回调（逆序执行）；须在 Init 前调用。Game 不解析回调语义。
func (g *Game) RegisterShutdownHook(fn func()) {
	if fn == nil {
		return
	}
	g.shutdownHooks = append(g.shutdownHooks, fn)
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

	//2.5 初始化路由

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
	g.urm = NewUserRequestManager(g.processUserRequest)
	//启动各个组件
	g.startConnectionServices()
	// 4.等待系统信号
	xos.WaitSysSignal(func(s os.Signal) {
		xlog.Info().Msgf("Received signal: %s, shutting down server...", s.String())
	})

	// 5. 释放资源
	g.shutdown()
}

// 启动各个组件
func (g *Game) startConnectionServices() error {
	g.wsServer = melody.New()
	//1.监听游戏消息
	g.listenGameMessage()
	//2.配置webSocket服务
	g.configureWebSocket()
	//3.配置http服务
	g.configureHttpServer()
	//4.启动http服务器
	if err := <-g.startHttpServer(); err != nil {
		xlog.Error().Err(err).Msg("http server start failed")
		return err
	}
	xlog.Info().Msgf("http server started on %s successfully", config.ApiPort())
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

// 关闭服务
func (g *Game) shutdown() {

	// 1. 创建一个带超时的上下文用于关闭
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 2. 关闭HTTP服务器
	if err := g.httpServer.Shutdown(shutdownCtx); err != nil {
		xlog.Error().Err(err).Msg("Failed to shutdown HTTP server")
	}

	// 3. 关闭WebSocket连接
	if err := g.wsServer.Close(); err != nil {
		xlog.Error().Err(err).Msg("Failed to close WebSocket server")
	}

	// 6. 关闭用户请求管理器
	g.urm.Close()

	for i := len(g.shutdownHooks) - 1; i >= 0; i-- {
		g.shutdownHooks[i]()
	}

	// 7. 收发器关闭
	if err := g.transceiver.Close(); err != nil {
		xlog.Error().Err(err).Msg("Failed to close transceiver")
	}

	xlog.Info().Msg("Server shutdown completed")
}

// 从路由表分发到具体业务处理器。
func (g *Game) processUserRequest(event *requestEvent) {
	header := event.data.GetHeader()
	if header == nil {
		xlog.Error().Msg("[processUserRequest] missing header")
		return
	}
	key := fmt.Sprintf("%s:%s", event.data.GetHeader().GetVersion(), event.data.GetRoute())
	handler, ok := g.routes.Load(key)
	if !ok {
		xlog.Error().Msgf("[processUserRequest] route %s not found", key)
		return
	}
	if err := handler.(GameMessageHandler)(g, event.s, event.data); err != nil {
		xlog.Error().Msgf("[processUserRequest] handler error: %v", err)
		return
	}
}
