package room

import (
	"errors"
	"strconv"
	"sync"
	"time"
)

var (
	// ErrRoomClosed 表示房间 Actor 已关闭，不再接受投递。
	ErrRoomClosed = errors.New("room actor is closed")
	// ErrMailboxFull 表示邮箱已满，无法非阻塞入队。
	ErrMailboxFull = errors.New("room actor mailbox is full")
	// ErrNoResolver 表示未配置房间解析器。
	ErrNoResolver = errors.New("room resolver is not configured")
	// ErrNoService 表示未配置业务服务。
	ErrNoService = errors.New("room service is not configured")
)

// Event 表示单房间串行管道中的最小消息单元。
type Event struct {
	RoomID    string    // 目标房间标识
	UID       int64     // 发起用户 ID
	Route     string    // 业务路由或消息类型
	Payload   []byte    // 原始请求体
	MsgID     string    // 消息去重或追踪标识
	EnqueueAt time.Time // 入队时刻，用于观测与延迟
}

// EventHandler 在房间 Actor 的 goroutine 中串行执行。
type EventHandler func(event Event)

// Actor 为单个房间提供串行队列，保证同房间内事件顺序处理。
type Actor struct {
	id      string        // 房间唯一标识
	handler EventHandler  // 非空时逐条处理邮箱中的事件
	mailbox chan Event    // 带缓冲的事件队列
	done    chan struct{} // 关闭时通知 run 已结束

	mu     sync.RWMutex // 保护 closed 与并发关闭
	closed bool         // 已关闭则拒绝新投递
	once   sync.Once    // 保证关闭逻辑只执行一次
}

// 创建实例并启动后台协程，从邮箱拉取并处理事件。
// id: 房间标识。
// mailboxSize: 邮箱容量；满时投递返回 ErrMailboxFull。
// handler: 非空时对每条事件调用。
// 返回: 已启动的后台处理实例指针。
func NewActor(id string, mailboxSize int, handler EventHandler) *Actor {
	a := &Actor{
		id:      id,
		handler: handler,
		mailbox: make(chan Event, mailboxSize),
		done:    make(chan struct{}),
	}
	go a.run()
	return a
}

// 将事件非阻塞投递到邮箱；关闭或已满时返回对应错误。
// event: 待处理的事件。
// 返回: nil 表示已入队，否则为 ErrRoomClosed 或 ErrMailboxFull。
func (a *Actor) Submit(event Event) error {
	a.mu.RLock()
	closed := a.closed
	a.mu.RUnlock()
	if closed {
		return ErrRoomClosed
	}
	select {
	case a.mailbox <- event:
		return nil
	default:
		return ErrMailboxFull
	}
}

// 关闭邮箱并阻塞至处理协程退出；可安全重复调用。
func (a *Actor) Close() {
	a.once.Do(func() {
		a.mu.Lock()
		a.closed = true
		close(a.mailbox)
		a.mu.Unlock()
		<-a.done
	})
}

// 从邮箱顺序读取事件并调用 handler，邮箱关闭后退出。
func (a *Actor) run() {
	defer close(a.done)
	for event := range a.mailbox {
		if a.handler != nil {
			a.handler(event)
		}
	}
}

// ManagerOptions 配置房间管理器创建 Actor 时的默认参数。
type ManagerOptions struct {
	MailboxSize int          // 每房间邮箱缓冲长度；<=0 时用默认值
	Handler     EventHandler // 新建 Actor 时使用的事件回调
}

// Manager 按房间 ID 懒创建并持有多个 Actor。
type Manager struct {
	mu      sync.RWMutex      // 保护 actors 映射
	actors  map[string]*Actor // 房间 ID 到 Actor
	options ManagerOptions    // 懒创建 Actor 时的配置
}

// 创建管理器；当 MailboxSize<=0 时默认 128。
func NewManager(options ManagerOptions) *Manager {
	if options.MailboxSize <= 0 {
		options.MailboxSize = 128
	}
	return &Manager{
		actors:  make(map[string]*Actor),
		options: options,
	}
}

// 解析或创建目标房间 Actor 后投递事件。
// roomID: 目标房间。
// event: 事件内容。
// 返回: 与底层 Actor 投递结果一致。
func (m *Manager) Submit(roomID string, event Event) error {
	actor := m.GetOrCreate(roomID)
	return actor.Submit(event)
}

// 若房间已存在则返回，否则懒创建并登记。
// roomID: 房间标识。
// 返回: 对应该房间的 Actor。
func (m *Manager) GetOrCreate(roomID string) *Actor {
	m.mu.RLock()
	actor, ok := m.actors[roomID]
	m.mu.RUnlock()
	if ok {
		return actor
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	actor, ok = m.actors[roomID]
	if ok {
		return actor
	}
	actor = NewActor(roomID, m.options.MailboxSize, m.options.Handler)
	m.actors[roomID] = actor
	return actor
}

// 关闭并清空所有房间 Actor，并等待各自退出。
func (m *Manager) Close() {
	m.mu.Lock()
	actors := m.actors
	m.actors = make(map[string]*Actor)
	m.mu.Unlock()

	for _, actor := range actors {
		actor.Close()
	}
}

// Resolver 根据用户与请求内容解析目标房间 ID。
type Resolver interface {
	Resolve(uid int64, route string, payload []byte) (string, error)
}

// UIDResolver 启动策略：以用户 ID 的十进制字符串作为房间 ID。
// 接入真实房间服务后可替换为 uid→roomId 查询。
type UIDResolver struct{}

// 构造默认的 UID 解析器实例。
func NewUIDResolver() *UIDResolver {
	return &UIDResolver{}
}

// 将 uid 格式化为十进制字符串作为房间 ID；忽略路由与负载。
// uid: 用户标识。
// route、payload: 预留，当前未使用。
// 返回: 房间 ID 与 nil 错误。
func (r *UIDResolver) Resolve(uid int64, route string, payload []byte) (string, error) {
	return strconv.FormatInt(uid, 10), nil
}

// Service 在房间串行上下文中执行业务逻辑。
type Service interface {
	Handle(uid int64, route string, payload []byte, msgID string)
}

// PipelineOptions 组装解析器、业务服务与邮箱等配置。
type PipelineOptions struct {
	MailboxSize int      // 每房间邮箱容量；<=0 时由 Manager 默认
	Resolver    Resolver // 必填，解析 uid/请求 到房间 ID
	Service     Service  // 必填，处理房间事件对应的业务
}

// Pipeline 连接解析、房间串行队列与业务服务。
type Pipeline struct {
	manager  *Manager // 按房间管理 Actor
	resolver Resolver // 解析目标房间
	service  Service  // 业务处理实现
}

// 在单处初始化房间层运行时对象；缺少解析器或服务时返回错误。
// options: 管道配置；Resolver 与 Service 不可为 nil。
// 返回: 已就绪的管道实例，或错误。
func Init(options PipelineOptions) (*Pipeline, error) {
	if options.Resolver == nil {
		return nil, ErrNoResolver
	}
	if options.Service == nil {
		return nil, ErrNoService
	}
	p := &Pipeline{
		resolver: options.Resolver,
		service:  options.Service,
	}
	p.manager = NewManager(ManagerOptions{
		MailboxSize: options.MailboxSize,
		Handler:     p.onRoomEvent,
	})
	return p, nil
}

// 解析房间后将请求封装为事件并投递到对应房间队列。
// uid、route、payload、msgID: 与业务请求一致。
// 返回: 解析或投递失败时的错误。
func (p *Pipeline) Dispatch(uid int64, route string, payload []byte, msgID string) error {
	roomID, err := p.resolver.Resolve(uid, route, payload)
	if err != nil {
		return err
	}
	return p.manager.Submit(roomID, Event{
		RoomID:    roomID,
		UID:       uid,
		Route:     route,
		Payload:   payload,
		MsgID:     msgID,
		EnqueueAt: time.Now(),
	})
}

// 主动创建房间 Actor，用于预热或提前占用资源。
// roomID: 要创建的房间标识。
func (p *Pipeline) CreateRoom(roomID string) {
	p.manager.GetOrCreate(roomID)
}

// 关闭底层管理器并释放所有房间 Actor。
func (p *Pipeline) Close() {
	p.manager.Close()
}

// 将房间事件转交给业务服务，在房间 goroutine 中调用。
// event: 已从队列取出并待处理的事件。
func (p *Pipeline) onRoomEvent(event Event) {
	p.service.Handle(event.UID, event.Route, event.Payload, event.MsgID)
}
