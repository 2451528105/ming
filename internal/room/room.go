package room

import (
	"errors"
	"strconv"
	"sync"
	"time"
)

var (
	ErrRoomClosed  = errors.New("room actor is closed")
	ErrMailboxFull = errors.New("room actor mailbox is full")
	ErrNoResolver  = errors.New("room resolver is not configured")
	ErrNoService   = errors.New("room service is not configured")
)

// Event is the minimal room-serialized message.
type Event struct {
	RoomID    string
	UID       int64
	Route     string
	Payload   []byte
	MsgID     string
	EnqueueAt time.Time
}

// EventHandler runs in room actor goroutine.
type EventHandler func(event Event)

// Actor keeps single-threaded processing for one room.
type Actor struct {
	id      string
	handler EventHandler
	mailbox chan Event
	done    chan struct{}

	mu     sync.RWMutex
	closed bool
	once   sync.Once
}

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

func (a *Actor) Close() {
	a.once.Do(func() {
		a.mu.Lock()
		a.closed = true
		close(a.mailbox)
		a.mu.Unlock()
		<-a.done
	})
}

func (a *Actor) run() {
	defer close(a.done)
	for event := range a.mailbox {
		if a.handler != nil {
			a.handler(event)
		}
	}
}

type ManagerOptions struct {
	MailboxSize int
	Handler     EventHandler
}

type Manager struct {
	mu      sync.RWMutex
	actors  map[string]*Actor
	options ManagerOptions
}

func NewManager(options ManagerOptions) *Manager {
	if options.MailboxSize <= 0 {
		options.MailboxSize = 128
	}
	return &Manager{
		actors:  make(map[string]*Actor),
		options: options,
	}
}

func (m *Manager) Submit(roomID string, event Event) error {
	actor := m.GetOrCreate(roomID)
	return actor.Submit(event)
}

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

func (m *Manager) Close() {
	m.mu.Lock()
	actors := m.actors
	m.actors = make(map[string]*Actor)
	m.mu.Unlock()

	for _, actor := range actors {
		actor.Close()
	}
}

// Resolver resolves which room should consume this player's request.
type Resolver interface {
	Resolve(uid int64, route string, payload []byte) (string, error)
}

// UIDResolver is a bootstrap strategy: one user maps to one room.
// Replace this with uid->roomId lookup when room service is implemented.
type UIDResolver struct{}

func NewUIDResolver() *UIDResolver {
	return &UIDResolver{}
}

func (r *UIDResolver) Resolve(uid int64, route string, payload []byte) (string, error) {
	return strconv.FormatInt(uid, 10), nil
}

// Service is the business layer executed by room actor.
type Service interface {
	Handle(uid int64, route string, payload []byte, msgID string)
}

type PipelineOptions struct {
	MailboxSize int
	Resolver    Resolver
	Service     Service
}

// Pipeline is the handler->room->service bridge.
type Pipeline struct {
	manager  *Manager
	resolver Resolver
	service  Service
}

// Init constructs room-layer runtime objects in one place.
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

// Dispatch resolves room and enqueues request event.
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

// CreateRoom eagerly creates room actor.
func (p *Pipeline) CreateRoom(roomID string) {
	p.manager.GetOrCreate(roomID)
}

func (p *Pipeline) Close() {
	p.manager.Close()
}

func (p *Pipeline) onRoomEvent(event Event) {
	p.service.Handle(event.UID, event.Route, event.Payload, event.MsgID)
}
