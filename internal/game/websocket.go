package game

import (
	"ming/internal/config"
	"ming/sdk/xlog"
	"strconv"

	"github.com/ivy-mobile/odin/envelope"
	"github.com/olahol/melody"
	"google.golang.org/protobuf/proto"
)

// 配置webSocket服务
func (g *Game) configureWebSocket() error {
	g.wsServer.Config.ConcurrentMessageHandling = false // 拒绝并发处理消息
	g.wsServer.Config.MaxMessageSize = 1024             // 最大消息大小
	g.handleConnect()
	g.handleDisconnect()
	g.handleMessage()
	return nil
}

func (g *Game) handleConnect() {
	g.wsServer.HandleConnect(func(s *melody.Session) {
		temp := s.Request.FormValue(UserKey)
		userId, err := strconv.ParseInt(temp, 10, 64)
		if err != nil {
			s.Close()
			return
		}
		if userId <= 0 {
			xlog.Error().Msgf("[Connect]invalid uid: %s", temp)
			_ = s.Write([]byte("invalid uid"))
			_ = s.Close()
			return
		}
		connID := g.idGen.NextID()
		// 保存到 melody.Session，便于断线时直接取 uid / connId
		s.Set(UserKey, userId)
		s.Set(ConnIDKey, connID)
		g.sessions.Store(userId, s)
		// 绑定玩家所在网关节点和连接ID，供跨节点路由查询
		if err = g.locator.BindGateNode(userId, g.nodeId, connID); err != nil {
			g.sessions.Delete(userId)
			xlog.Error().Err(err).Msgf("[Connect]bind gate node failed, uid: %d, connId: %d", userId, connID)
			_ = s.Write([]byte("bind gate node failed"))
			_ = s.Close()
			return
		}
		xlog.Info().Msgf("[Connect]user %d connected, connId: %d, ip: %s", userId, connID, s.Request.RemoteAddr)

	})
}

func (g *Game) handleDisconnect() {
	g.wsServer.HandleDisconnect(func(s *melody.Session) {
		uidRaw, ok := s.Get(UserKey)
		if !ok {
			xlog.Info().Msg("[Disconnect]missing uid in session, skip cleanup")
			return
		}
		userId, ok := uidRaw.(int64)
		if !ok {
			xlog.Error().Msg("[Disconnect]invalid uid type in session")
			return
		}

		connRaw, ok := s.Get(ConnIDKey)
		if !ok {
			xlog.Info().Msgf("[Disconnect]missing connId, uid: %d", userId)
			return
		}
		connID, ok := connRaw.(int64)
		if !ok {
			xlog.Error().Msgf("[Disconnect]invalid connId type, uid: %d", userId)
			return
		}

		currentRaw, exists := g.sessions.Load(userId)
		if !exists {
			return
		}
		current, ok := currentRaw.(*melody.Session)
		if !ok {
			xlog.Error().Msgf("[Disconnect]invalid state type, uid: %d", userId)
			return
		}
		if current != s {
			// 旧连接晚到的断线事件：不动当前在线映射，避免误踢当前连接
			xlog.Info().Msgf("[Disconnect]stale session closed, uid: %d, connId: %d", userId, connID)
			return
		}
		g.sessions.Delete(userId)
		if err := g.locator.UnbindGateNode(userId, g.nodeId, connID); err != nil {
			xlog.Error().Err(err).Msgf("[Disconnect]unbind gate node failed, uid: %d, connId: %d", userId, connID)
		}
		xlog.Info().Msgf("[Disconnect]user %d disconnected, connId: %d", userId, connID)

	})
}

func (g *Game) handleMessage() {
	// 处理二进制消息
	g.wsServer.HandleMessageBinary(func(s *melody.Session, msg []byte) {
		var data envelope.InputMessage
		if err := proto.Unmarshal(msg, &data); err != nil {
			xlog.Error().Msgf("[handleRequestProtoMessage] proto.Unmarshal error: %v", err)
			return
		}
		u, ok := s.Get(UserKey)
		header := data.GetHeader()
		if !ok || header.GetUid() != u.(int64) {
			xlog.Error().Msgf("[handleRequestProtoMessage] 非法请求, 传入的Uid与会话绑定的playerId不匹配, uid: %d, playerId: %v", header.GetUid(), u)
			return
		}

		// 方式2: 入列用户请求管理器 - 异步处理，一个用户一个协程，用户与用户之间互不阻塞 - 推荐
		// g.wsServer.Config.ConcurrentMessageHandling 值为 true 或 false 都可以，具体差异待观察
		g.urm.Go(int(header.GetUid()), &requestEvent{
			s:    s,
			data: &data,
		})
	})
}

// SendMessageByRMQ 发送消息 - 通过消息队列
func (g *Game) SendMessageByRMQ(uid int64, data []byte) {

	node, err := g.locator.GetGateNode(uid)
	if err != nil {
		xlog.Error().Msgf("[SendMessageByRMQ] GetGateNode error: %v, uid: %v", err, uid)
		return
	}
	if node == "" {
		xlog.Error().Msgf("[SendMessageByRMQ] GetGateNode node is empty, uid: %v", uid)
		return
	}
	msgId, err := g.transceiver.SendMessage(uid, data, config.Cfg.Application.Name, node)
	if err != nil {
		xlog.Error().Msgf("[Transceiver] SendMessage error: %v, uid: %v, node: %v, msgId: %v", err, uid, node, msgId)
		return
	}
}
