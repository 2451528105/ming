package game

import (
	"ming/sdk/xlog"
	"strconv"

	"github.com/olahol/melody"
)

// 配置webSocket服务
func (g *Game) configureWebSocket() error {
	g.wsServer.Config.ConcurrentMessageHandling = false // 拒绝并发处理消息
	g.wsServer.Config.MaxMessageSize = 1024             // 最大消息大小
	g.handleConnect()
	g.handleDisconnect()

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
		xlog.Info().Msgf("[Disconnect]user %d disconnected, connId: %d", userId, connID)

	})
}
