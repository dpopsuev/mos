package store

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

type Session struct {
	ID            string
	WorkspaceHash string
	CreatedAt     time.Time
}

type sessionEntry struct {
	session       Session
	workspaceHash string
}

type storeHandle struct {
	store    *BoltStore
	lastUsed time.Time
	refCount int
}

type SessionManager struct {
	mu          sync.Mutex
	sessions    map[string]*sessionEntry
	stores      map[string]*storeHandle // workspaceHash -> handle
	idleTimeout time.Duration
	stopReaper  chan struct{}
}

func NewSessionManager(idleTimeout time.Duration) *SessionManager {
	m := &SessionManager{
		sessions:    make(map[string]*sessionEntry),
		stores:      make(map[string]*storeHandle),
		idleTimeout: idleTimeout,
		stopReaper:  make(chan struct{}),
	}
	go m.reapLoop()
	return m
}

func (m *SessionManager) Connect(workspaceRoots []string) (Session, error) {
	hash := WorkspaceHash(workspaceRoots)

	m.mu.Lock()
	defer m.mu.Unlock()

	h, ok := m.stores[hash]
	if !ok {
		dbPath := dbPathForHash(hash)
		s, err := Open(dbPath)
		if err != nil {
			return Session{}, fmt.Errorf("open store for workspace %s: %w", hash, err)
		}
		h = &storeHandle{store: s.(*BoltStore), lastUsed: time.Now()}
		m.stores[hash] = h
	}
	h.refCount++
	h.lastUsed = time.Now()

	sid := generateSessionID()
	sess := Session{ID: sid, WorkspaceHash: hash, CreatedAt: time.Now()}
	m.sessions[sid] = &sessionEntry{session: sess, workspaceHash: hash}
	return sess, nil
}

func (m *SessionManager) Disconnect(sessionID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	entry, ok := m.sessions[sessionID]
	if !ok {
		return
	}
	delete(m.sessions, sessionID)

	if h, ok := m.stores[entry.workspaceHash]; ok {
		h.refCount--
	}
}

func (m *SessionManager) Route(sessionID string) (Store, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	entry, ok := m.sessions[sessionID]
	if !ok {
		return nil, fmt.Errorf("unknown session: %s", sessionID)
	}
	h, ok := m.stores[entry.workspaceHash]
	if !ok {
		return nil, fmt.Errorf("store closed for workspace: %s", entry.workspaceHash)
	}
	h.lastUsed = time.Now()
	return h.store, nil
}

func (m *SessionManager) ActiveSessions() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.sessions)
}

func (m *SessionManager) Close() error {
	close(m.stopReaper)
	m.mu.Lock()
	defer m.mu.Unlock()

	var firstErr error
	for hash, h := range m.stores {
		if err := h.store.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
		delete(m.stores, hash)
	}
	m.sessions = make(map[string]*sessionEntry)
	return firstErr
}

func (m *SessionManager) reapLoop() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-m.stopReaper:
			return
		case <-ticker.C:
			m.reapIdle()
		}
	}
}

func (m *SessionManager) reapIdle() {
	m.mu.Lock()
	defer m.mu.Unlock()

	cutoff := time.Now().Add(-m.idleTimeout)
	for hash, h := range m.stores {
		if h.refCount == 0 && h.lastUsed.Before(cutoff) {
			h.store.Close()
			delete(m.stores, hash)
		}
	}
}

func WorkspaceHash(roots []string) string {
	if len(roots) == 0 {
		cwd, _ := os.Getwd()
		roots = []string{cwd}
	}
	sorted := make([]string, len(roots))
	copy(sorted, roots)
	sort.Strings(sorted)
	h := sha256.Sum256([]byte(strings.Join(sorted, "\n")))
	return fmt.Sprintf("%x", h[:6])
}

func DefaultSocketPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".mosbus", "mosbus.sock")
}

func DefaultPIDPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".mosbus", "mosbus.pid")
}

func dbPathForHash(hash string) string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".mosbus", hash, "store.db")
}

func generateSessionID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}
