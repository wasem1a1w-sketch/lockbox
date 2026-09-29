package server

import (
	"sync"
	"time"
)

// session holds the unlocked vault key in memory only.
// Never persisted, never logged. Cleared by lock, idle timeout, or vault-path change.
type session struct {
	mu         sync.Mutex
	key        []byte
	salt       []byte
	iterations int
	vaultPath  string
	lastActive time.Time
}

func (s *session) touch() {
	s.mu.Lock()
	s.lastActive = time.Now()
	s.mu.Unlock()
}

type sessionManager struct {
	mu      sync.Mutex
	current *session
	timeout time.Duration
	stop    chan struct{}
	once    sync.Once
}

func newSessionManager(timeout time.Duration) *sessionManager {
	return &sessionManager{timeout: timeout, stop: make(chan struct{})}
}

func (m *sessionManager) set(key, salt []byte, iterations int, vaultPath string) {
	m.mu.Lock()
	m.current = &session{
		key:        key,
		salt:       salt,
		iterations: iterations,
		vaultPath:  vaultPath,
		lastActive: time.Now(),
	}
	m.mu.Unlock()
}

// get returns the live session, enforcing idle timeout and vault-path identity.
func (m *sessionManager) get(vaultPath string) *session {
	m.mu.Lock()
	s := m.current
	m.mu.Unlock()
	if s == nil {
		return nil
	}
	s.mu.Lock()
	idle := time.Since(s.lastActive)
	pathMatches := s.vaultPath == vaultPath
	s.mu.Unlock()
	if idle > m.timeout || !pathMatches {
		m.clear()
		return nil
	}
	return s
}

func (m *sessionManager) clear() {
	m.mu.Lock()
	m.current = nil
	m.mu.Unlock()
}

func (m *sessionManager) active() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.current != nil
}

func (m *sessionManager) startJanitor() {
	m.once.Do(func() {
		go func() {
			ticker := time.NewTicker(30 * time.Second)
			defer ticker.Stop()
			for {
				select {
				case <-m.stop:
					return
				case <-ticker.C:
					m.mu.Lock()
					s := m.current
					m.mu.Unlock()
					if s == nil {
						continue
					}
					s.mu.Lock()
					idle := time.Since(s.lastActive)
					s.mu.Unlock()
					if idle > m.timeout {
						m.clear()
					}
				}
			}
		}()
	})
}

func (m *sessionManager) stopJanitor() {
	m.once.Do(func() {})
	select {
	case <-m.stop:
	default:
		close(m.stop)
	}
}
