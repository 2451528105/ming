package room

import (
	"sync"

	"ming/internal/service"
)

// ManagerOptions 只控制为新房间分配的邮箱容量。
type ManagerOptions struct {
	MailboxSize int
}

// Manager 持有 roomID → *Room。
type Manager struct {
	mu      sync.Mutex
	rooms   map[string]*Room
	options ManagerOptions
}

func NewManager(options ManagerOptions) *Manager {
	if options.MailboxSize <= 0 {
		options.MailboxSize = 128
	}
	return &Manager{
		rooms:   make(map[string]*Room),
		options: options,
	}
}

func (m *Manager) Register(roomID string) error {
	if roomID == "" {
		return ErrInvalidRoomID
	}
	m.GetOrCreate(roomID)
	return nil
}

func (m *Manager) GetOrCreate(roomID string) *Room {
	m.mu.Lock()
	defer m.mu.Unlock()
	if r, ok := m.rooms[roomID]; ok {
		return r
	}
	r := NewRoom(roomID, m.options.MailboxSize)
	m.rooms[roomID] = r
	return r
}

func (m *Manager) Submit(roomID string, e *Event) error {
	r := m.GetOrCreate(roomID)
	return r.Submit(e)
}

func (m *Manager) CloseRoom(roomID string) {
	m.mu.Lock()
	r := m.rooms[roomID]
	delete(m.rooms, roomID)
	m.mu.Unlock()
	if r != nil {
		r.Close()
	}
}

func (m *Manager) Close() {
	m.mu.Lock()
	list := m.rooms
	m.rooms = make(map[string]*Room)
	m.mu.Unlock()

	for _, r := range list {
		r.Close()
	}
}

var defaultMgr *Manager

func InitDefault() error {
	if defaultMgr != nil {
		return nil
	}
	defaultMgr = NewManager(ManagerOptions{MailboxSize: 256})
	return nil
}

// CreateRoom 先向 service 申请房间号，再在默认 Manager 下挂载运行时。
func CreateRoom() (roomID string, err error) {
	if defaultMgr == nil {
		return "", ErrNotInstalled
	}
	roomID, err = service.RoomCreate()
	if err != nil {
		return "", err
	}
	if err := defaultMgr.Register(roomID); err != nil {
		return "", err
	}
	return roomID, nil
}

// Register 已知 roomID 时仅挂载运行时（不参与发号）。
func Register(roomID string) error {
	if defaultMgr == nil {
		return ErrNotInstalled
	}
	return defaultMgr.Register(roomID)
}

func Submit(roomID string, e *Event) error {
	if defaultMgr == nil {
		return ErrNotInstalled
	}
	return defaultMgr.Submit(roomID, e)
}

func DefaultManager() *Manager {
	return defaultMgr
}
