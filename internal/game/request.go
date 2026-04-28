package game

import (
	"context"
	"ming/internal/pb"
	"ming/sdk/xlog"
	"sync"

	"github.com/olahol/melody"
	"golang.org/x/time/rate"
)

// 请求事件
type requestEvent struct {
	s    *melody.Session
	data *pb.InputMessage
}

// UserRequestManager 用户请求管理器
// 一个用户一个协程(一个执行队列)保证用户维度请求处理同步
// 可设置限流器，控制用户请求频率
type UserRequestManager struct {
	msgHandler    func(event *requestEvent)
	requestQueues sync.Map // 用户请求队列 key: playerId, value: chan *requestEvent
	limiters      sync.Map // 用户请求限流器 key: playerId, value: *rate.Limiter

	// 添加上下文和取消函数的存储键
	ctxs    sync.Map // 用户上下文 key: playerId, value: context.Context
	cancels sync.Map // 用户取消函数 key: playerId, value: context.CancelFunc
}

func NewUserRequestManager(msgHandler func(event *requestEvent)) *UserRequestManager {
	return &UserRequestManager{
		msgHandler: msgHandler,
	}
}

// Go 请求事件入列
func (u *UserRequestManager) Go(uid int, event *requestEvent) {

	limiter, _ := u.limiters.LoadOrStore(uid, rate.NewLimiter(5, 10))

	if !limiter.(*rate.Limiter).Allow() {

		xlog.Warn().Int("Player", uid).Msg("too many request ! ")

		if err := event.s.Write([]byte("too many requests")); err != nil {
			xlog.Error().Err(err).Int("Player", uid).Msg("write too many requests error")
		}
		return
	}

	queue, _ := u.requestQueues.LoadOrStore(uid, make(chan *requestEvent, 10))

	// 第一次请求 启动协程处理队列中的请求
	if _, ok := u.ctxs.Load(uid); !ok {
		ctx, cancel := context.WithCancel(context.Background())
		u.ctxs.Store(uid, ctx)
		u.cancels.Store(uid, cancel)

		go func() {
			defer func() {
				// 协程退出时清理资源
				close(queue.(chan *requestEvent))
				u.ctxs.Delete(uid)
				u.cancels.Delete(uid)
				u.requestQueues.Delete(uid)
				u.limiters.Delete(uid)
			}()

			for {
				select {
				case <-ctx.Done():
					xlog.Info().Int("Player", uid).Msgf("UserRequestManager context Done! uid: %d", uid)
					return
				case e := <-queue.(chan *requestEvent):
					u.msgHandler(e)
				}
			}
		}()
	}

	// 入列时,检查通道是否已关闭，避免panic
	select {
	case queue.(chan *requestEvent) <- event:
	default:
		xlog.Warn().Int("Uid", uid).Msg("request queue is full")
		if err := event.s.Write([]byte("request queue is full")); err != nil {
			xlog.Error().Err(err).Int("Uid", uid).Msg("write queue full error")
		}
	}
}

// TickUser 踢出用户，释放对应资源
// 玩家长时间离线时需要调用此方法，释放资源
func (u *UserRequestManager) TickUser(uid int64) {
	if cancel, ok := u.cancels.Load(uid); ok {
		cancel.(context.CancelFunc)()
	}
}

// Close 关闭所有用户请求管理器
func (u *UserRequestManager) Close() {
	u.cancels.Range(func(key, value interface{}) bool {
		cancel := value.(context.CancelFunc)
		cancel()
		return true
	})
	xlog.Info().Msg("UserRequestManager stopped")
}
