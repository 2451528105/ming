package game

import (
	"fmt"
	"ming/internal/pb"
	"ming/sdk/consts"
	"ming/sdk/xlog"
	"time"

	"github.com/olahol/melody"
	"google.golang.org/protobuf/proto"
)

type Context interface {
	Seq() uint64
	Uid() int64
	Route() string
	GameId() int32
	MsgId() string
	Timestamp() int64
	Version() string

	Cost() time.Duration // 耗时 = 当前时间与创建时间之间的时间差
	OkResp(...proto.Message)
	ErrResp(consts.ErrorCode, ...string)
	Push(msgTag string, data proto.Message, uids ...int64)
}
type defaultContext struct {
	g          *Game
	session    *melody.Session        // 当前会话
	req        *pb.RequestMessage     // 本次请求数据
	createTime int64                  // 进入请求时间 ms
}

func newDefaultContext(g *Game, s *melody.Session, req *pb.RequestMessage) *defaultContext {
	return &defaultContext{
		g:          g,
		session:    s,
		req:        req,
		createTime: time.Now().UnixMilli(),
	}
}

// validateEnvelope 校验信封/路由等传输层必填字段，不包含业务规则（如 gameId 是否为本产品由 handler 决定）。
func (d *defaultContext) validateEnvelope() bool {
	req := d.req
	if req == nil {
		d.ErrResp(consts.ErrorCode_RequestErr, "invalid request")
		return false
	}
	header := req.GetHeader()
	if header.GetMsgId() == "" {
		d.ErrResp(consts.ErrorCode_RequestErr, "invalid msgId")
		return false
	}
	if header.GetUid() == 0 {
		d.ErrResp(consts.ErrorCode_RequestErr, "invalid uid")
		return false
	}
	if header.GetGameId() == 0 {
		d.ErrResp(consts.ErrorCode_RequestErr, "invalid gameId")
		return false
	}
	if header.GetTimestamp() == 0 {
		d.ErrResp(consts.ErrorCode_RequestErr, "invalid timestamp")
		return false
	}
	if header.GetVersion() == "" {
		d.ErrResp(consts.ErrorCode_RequestErr, "invalid version")
		return false
	}
	if req.GetRoute() == "" {
		d.ErrResp(consts.ErrorCode_RequestErr, "invalid route")
		return false
	}
	return true
}
func (d *defaultContext) ErrResp(code consts.ErrorCode, msg ...string) {
	if d == nil {
		fmt.Println("context is nil")
		return
	}
	if d.session == nil {
		xlog.Error().Msg("session is nil")
		return
	}
	pm := &pb.ServerMessage{
		Header: &pb.Header{
			Seq:       d.Seq(),
			Uid:       d.Uid(),
			MsgId:     d.MsgId(),
			GameId:    d.GameId(),
			Timestamp: time.Now().UnixMilli(),
			Version:   d.Version(),
		},
		MsgTag:    d.Route(),
		MsgType:   consts.WsTag_Reply,
		ErrorCode: string(code),
	}
	if len(msg) > 0 {
		pm.ErrorMsg = msg[0]
	}
	bytes, _ := proto.Marshal(pm)

	d.g.SendMessageByRMQ(d.Uid(), bytes) // 发送消息 - 通过消息队列

	xlog.Error().
		Int("Player", int(d.Uid())).
		Str("Route", d.Route()).
		Str("MsgId", d.MsgId()).
		Str("ErrCode", string(code)).Str("Tip", pm.ErrorMsg).
		Msgf("response error! cost: %v", d.Cost())
}

// Push 推送消息
func (d *defaultContext) Push(msgTag string, msg proto.Message, uids ...int64) {
	if len(uids) <= 0 {
		xlog.Info().Msgf("[Push] canceled ! msgTag: %v, uids: %v, cost: %v", msgTag, uids, d.Cost())
		return
	}
	data, _ := proto.Marshal(msg)
	for _, uid := range uids {
		pm := &pb.ServerMessage{
			Header: &pb.Header{
				Uid:       uid,
				GameId:    d.GameId(),
				Version:   d.Version(),
				Timestamp: time.Now().UnixMilli(),
			},
			MsgType:   consts.WsTag_Push,
			ErrorCode: string(consts.ErrorCode_Success),
			MsgTag:    msgTag,
			Data:      data,
		}
		bytes, _ := proto.Marshal(pm)
		d.g.SendMessageByRMQ(uid, bytes)
	}
	xlog.Info().Msgf("[Push] success ! msgTag: %v, uids: %v, cost: %v", msgTag, uids, d.Cost())
}

func (d *defaultContext) OkResp(ds ...proto.Message) {
	if d == nil {
		fmt.Println("context is nil")
		return
	}
	if d.session == nil {
		xlog.Error().Msg("session is nil")
		return
	}
	pm := &pb.ServerMessage{
		Header: &pb.Header{
			GameId:    d.GameId(),
			MsgId:     d.MsgId(),
			Seq:       d.Seq(),
			Uid:       d.Uid(),
			Timestamp: time.Now().UnixMilli(),
			Version:   d.Version(),
		},
		MsgTag:    d.Route(),
		MsgType:   consts.WsTag_Reply,
		ErrorCode: string(consts.ErrorCode_Success),
	}
	if len(ds) > 0 {
		data, _ := proto.Marshal(ds[0])
		pm.Data = data
	}
	bytes, _ := proto.Marshal(pm)

	d.g.SendMessageByRMQ(d.Uid(), bytes)

	xlog.Info().
		Int("Player", int(d.Uid())).
		Str("Route", d.Route()).
		Str("MsgId", d.MsgId()).
		Msgf("[Response] success! cost: %d ms", time.Now().UnixMilli()-d.createTime)

}

// -------------------------- 获取公共必选参数 --------------------------
func (d *defaultContext) Seq() uint64 {
	return d.req.GetHeader().GetSeq()
}
func (d *defaultContext) Uid() int64 {
	return d.req.GetHeader().GetUid()
}
func (d *defaultContext) GameId() int32 {
	return d.req.GetHeader().GetGameId()
}

func (d *defaultContext) Route() string {
	return d.req.GetRoute()
}
func (d *defaultContext) MsgId() string {
	return d.req.GetHeader().GetMsgId()
}

func (d *defaultContext) Timestamp() int64 {
	return d.req.GetHeader().GetTimestamp()
}

func (d *defaultContext) Version() string {
	return d.req.GetHeader().GetVersion()
}

func (d *defaultContext) Cost() time.Duration {
	return time.Millisecond * time.Duration(time.Now().UnixMilli()-d.createTime)
}

