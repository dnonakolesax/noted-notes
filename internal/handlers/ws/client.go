package ws

import (
	"time"

	"github.com/automerge/automerge-go"
	"github.com/fasthttp/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 8 << 20 // 8 MB
)

type Client struct {
	conn    *websocket.Conn
	ss      *automerge.SyncState
	send    chan OutMsg
	sess    *DocSession
	blockID string
	nbID    string
}

type OutMsg struct {
	Type int
	Data []byte
}

func newClient(conn *websocket.Conn) *Client {
	return &Client{
        conn: conn,
        send: make(chan OutMsg, 32),
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		_ = c.conn.Close()
	}()

	for {
		select {
		case out, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.conn.WriteMessage(out.Type, out.Data); err != nil {
				return
			}

		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) readPump(
	onBinary func([]byte),
	onText func([]byte),
) {
	defer func() { _ = c.conn.Close() }()

	c.conn.SetReadLimit(maxMessageSize)
	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		mt, msg, err := c.conn.ReadMessage()
		if err != nil {
			return
		}

		switch mt {
		case websocket.BinaryMessage:
			if onBinary != nil {
				onBinary(msg)
			}
		case websocket.TextMessage:
			if onText != nil {
				onText(msg)
			}
		default:
		}
	}
}
