package ws

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
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

    nid, blockID := splitDocID(safeID)
	
	blockID = strings.ReplaceAll(blockID, "_", "")
	nid = strings.ReplaceAll(nid, "_", "-")

	path := fmt.Sprintf("/noted/codes/kernels/%s/block_%s", nid, blockID)
	raw, err := os.ReadFile(path)
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

	hotFile := filepath.Join(hotDir, "block_"+blkSafe)

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

func (m *Manager) Release(ctx context.Context, s *DocSession) {
	m.mu.Lock()
	defer m.mu.Unlock()

    fileID, blockID := splitDocID(s.id)
	err := os.WriteFile("/noted/codes/" + fileID + "/block_" + blockID, []byte(s.doc.Path("text").String()), 0o777)
	if err != nil {
		slog.Error("error writing file after release: ", slog.String("error", err.Error()))
	}
	_ = s.store.Upload(ctx, blockID, s.doc.Save())
	s.watchers--
	if s.watchers > 0 {
		return
	}

	//_ = os.Remove(s.hotPath)

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
    if c.blocks == nil {
        return
    }

    for blockID, bp := range c.blocks {
        sess := bp.sess

        sess.mu.Lock()
        delete(sess.clients, c)
        sess.mu.Unlock()

        m.Release(ctx, sess) // уменьшит watchers и при 0 выгрузит блок из hot
        delete(c.blocks, blockID)
    }
}

func (m *Manager) Save(ctx context.Context, c *Client, blockID string) {
	// bl, ok := c.blocks[blockID]

	// if !ok {
	// 	fmt.Printf("block not found: %s", blockID)
	// 	return
	// }

	// s := bl.sess
	// _ = s.store.Upload(ctx, blockID, s.doc.Save())
	// _ = os.Remove(s.hotPath)
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
