package game

import (
	"fmt"

	"github.com/ivy-mobile/odin/envelope"
	"github.com/olahol/melody"
	"google.golang.org/protobuf/proto"
)

type (
	// GameMessageHandler 由上层注册；具体业务在 handler 包实现，此处仅为分发回调类型。
	GameMessageHandler func(g *Game, s *melody.Session, msg *envelope.InputMessage) error
)

func Handler[I any](fn func(Context, *I)) GameMessageHandler {
	return func(g *Game, s *melody.Session, msg *envelope.InputMessage) error {
		// 解析业务数据 payload 到传入的类型I
		var in I
		if len(msg.GetPayload()) > 0 {
			if err := proto.Unmarshal(msg.GetPayload(), any(&in).(proto.Message)); err != nil {
				return fmt.Errorf("[Handler] unmarshal payload faild: %w", err)
			}
		}
		if ctx := newDefaultContext(g, s, msg); ctx.validateEnvelope() {
			//xlog.Info().Int64("Player", ctx.Uid()).Str("Route", msg.GetRoute()).Str("MsgId", msg.GetMsgId()).Msgf("[Request] ... req: %s", in)
			fn(ctx, &in)
		}
		return nil
	}
}

// RegisterRoute 注册路由
func (g *Game) RegisterRoute(version, route string, handler GameMessageHandler) {
	key := fmt.Sprintf("%s:%s", version, route)
	g.routes.Store(key, handler)
}
