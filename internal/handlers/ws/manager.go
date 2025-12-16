package ws

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/automerge/automerge-go"
	"github.com/fasthttp/websocket"
)

type Store interface {
	Get(ctx context.Context, id string) ([]byte, error)
	Upload(ctx context.Context, id string, data []byte) error
}

type DocSession struct {
	id      string
	doc     *automerge.Doc
	store   Store
	hotPath string

	mu       sync.Mutex
	clients  map[*Client]struct{}
	watchers int
}

type Manager struct {
	mu     sync.Mutex
	sess   map[string]*DocSession
	hotDir string
	store  Store
}

func splitDocID(id string) (notebook, block string) {
	parts := strings.SplitN(id, "::", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return id, "root"
}

func NewManager(st Store, hotDir string) *Manager {
	return &Manager{
		sess:   make(map[string]*DocSession),
		hotDir: hotDir,
		store:  st,
	}
}

func sanitizeID(id string) string {
	safe := strings.TrimSpace(id)
	safe = strings.ReplaceAll(safe, "..", "")
	safe = strings.ReplaceAll(safe, "/", "_")
	safe = strings.ReplaceAll(safe, "\\", "_")
	if safe == "" {
		safe = "default"
	}
	return safe
}

func (m *Manager) Acquire(ctx context.Context, id string) (*DocSession, error) {
	safeID := sanitizeID(id)

	m.mu.Lock()
	if s, ok := m.sess[safeID]; ok {
		s.watchers++
		m.mu.Unlock()
		return s, nil
	}
	m.mu.Unlock()

    _, blockID := splitDocID(safeID)
	raw, err := m.store.Get(ctx, blockID)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}

	var doc *automerge.Doc
	if len(raw) == 0 {
		doc = automerge.New()
	} else {
		doc, err = automerge.Load(raw)
		if err != nil {
			return nil, err
		}
	}

	nb, blk := splitDocID(safeID)
	nbSafe := sanitizeID(nb) 
	blkSafe := sanitizeID(blk)

	hotDir := filepath.Join(m.hotDir, nbSafe)
	if err := os.MkdirAll(hotDir, 0o755); err != nil {
		return nil, err
	}

	hotFile := filepath.Join(hotDir, "block_"+blkSafe+".am")

	s := &DocSession{
		id:       safeID,
		doc:      doc,
		store:    m.store,
		hotPath:  hotFile,
		clients:  make(map[*Client]struct{}),
		watchers: 1,
	}
	_ = os.WriteFile(s.hotPath, doc.Save(), 0o644)

	m.mu.Lock()
	m.sess[safeID] = s
	m.mu.Unlock()

	return s, nil
}

func (m *Manager) SwitchBlock(ctx context.Context, c *Client, blockID string) (*DocSession, error) {
    // docID для блока: nb::block
    docID := c.nbID + "::" + blockID

    // если уже на этом блоке — ничего не делаем
    if c.sess != nil && c.sess.id == docID {
        return c.sess, nil
    }

    // отцепить от старого блока, если был
    if old := c.sess; old != nil {
        old.mu.Lock()
        delete(old.clients, c)
        old.mu.Unlock()

        m.Release(ctx, old) // уменьшит watchers и, если надо, выгрузит блок
    }

    // подключаемся к новому блоку
    sess, err := m.Acquire(ctx, docID)
    if err != nil {
        return nil, err
    }

    sess.mu.Lock()
    sess.clients[c] = struct{}{}
    sess.mu.Unlock()

    c.blockID = blockID
    c.sess = sess
    c.ss = automerge.NewSyncState(sess.doc)

    return sess, nil
}

func (m *Manager) Release(ctx context.Context, s *DocSession) {
	m.mu.Lock()
	defer m.mu.Unlock()

	s.watchers--
	if s.watchers > 0 {
		return
	}

    _, blockID := splitDocID(s.id)
	_ = s.store.Upload(ctx, blockID, s.doc.Save())
	_ = os.Remove(s.hotPath)

	delete(m.sess, s.id)
}

func (m *Manager) BroadcastNotebookText(nbID string, from *Client, data []byte) {
	prefix := nbID + "::"

	m.mu.Lock()
	sessions := make([]*DocSession, 0)
	for id, s := range m.sess {
		if strings.HasPrefix(id, prefix) {
			sessions = append(sessions, s)
		}
	}
	m.mu.Unlock()

	for _, s := range sessions {
		s.broadcastText(from, data)
	}
}

func (m *Manager) Disconnect(ctx context.Context, c *Client) {
    if c.sess == nil {
        return
    }

    s := c.sess
    s.mu.Lock()
    delete(s.clients, c)
    s.mu.Unlock()

    m.Release(ctx, s) // уменьшит watchers и, при 0, сохранит и удалит hot-файл
    c.sess = nil
}

func (s *DocSession) broadcastText(from *Client, msg []byte) {
	s.mu.Lock()
	clients := make([]*Client, 0, len(s.clients))
	for c := range s.clients {
		clients = append(clients, c)
	}
	s.mu.Unlock()

	for _, c := range clients {
		if from != nil && c == from {
			continue
		}
		select {
		case c.send <- OutMsg{Type: websocket.TextMessage, Data: msg}:
		default:
		}
	}
}
