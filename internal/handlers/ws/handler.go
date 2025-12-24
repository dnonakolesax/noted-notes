package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/automerge/automerge-go"
	"github.com/dnonakolesax/noted-notes/internal/consts"
	"github.com/dnonakolesax/noted-notes/internal/middleware"
	"github.com/fasthttp/router"
	"github.com/fasthttp/websocket"
	"github.com/valyala/fasthttp"
)

var upgrader = websocket.FastHTTPUpgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(ctx *fasthttp.RequestCtx) bool { return true },
}

type AccessService interface {
	Get(fileID string, userID string, byBlock bool) (string, error)
}

type SocketHandler struct {
	mgr           *Manager
	accessService AccessService
}

func NewHandler(mgr *Manager, accessService AccessService) *SocketHandler {
	return &SocketHandler{mgr: mgr, accessService: accessService}
}

func (sh *SocketHandler) Handle(ctx *fasthttp.RequestCtx) {
	docID := ctx.Request.URI().QueryArgs().Peek("nb")
	if docID == nil {
		// http.Error(w, "missing doc", http.StatusBadRequest)
		fmt.Println("missing doc")
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}
	rights, err := sh.accessService.Get(string(docID), ctx.UserValue(consts.CtxUserIDKey).(string), false)

	if err != nil {
		fmt.Println("no rights found", slog.String("fileid", string(docID)), slog.String("userid", ctx.UserValue(consts.CtxUserIDKey).(string)))
		ctx.SetStatusCode(fasthttp.StatusUnauthorized)
		return
	}

    if !strings.Contains(rights, "w") {
		fmt.Println("no rights for writing", slog.String("fileid", string(docID)), slog.String("userid", ctx.UserValue(consts.CtxUserIDKey).(string)))
		ctx.SetStatusCode(fasthttp.StatusUnauthorized)
		return
    }

	defer func() {
		res := recover()

		if res != nil {
			fmt.Printf("PANIC OCCURED: %v", res)
		}
	}()

	// sess, err := sh.mgr.Acquire(context.Background(), string(docID))
	// if err != nil {
	// 	// http.Error(w, err.Error(), http.StatusInternalServerError)
	// 	fmt.Printf("error session: %v \n", err)
	// 	ctx.SetStatusCode(fasthttp.StatusInternalServerError)
	// 	return
	// }
	// defer sh.mgr.Release(context.Background(), sess)
	//userID := r.URL.Query().Get("userId")
	//if userID == "" {
	//userID := "user_" + time.Now().Format("20060102150405")
	//}
	err = upgrader.Upgrade(ctx, func(conn *websocket.Conn) {
		client := newClient(conn)

		client.nbID = string(docID)
		go client.writePump()

		client.readPump(
			// onBinary (Automerge sync)
			func(msg []byte) {
				sh.handleBinary(client, msg)
			},

			// onText (presence)
			func(msg []byte) {
				sh.handleText(client, msg)
			},
		)
		sh.mgr.Disconnect(context.Background(), client)
	})

	if err != nil {
		fmt.Printf("error upgrade: %v \n", err)
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		return
	}
}

func (sh *SocketHandler) handleBinary(c *Client, msg []byte) {
	blockID, payload, ok := splitFrame(msg)
	if !ok {
		return
	}

	bp, err := sh.getOrAttachBlockPeer(c, blockID)
	if err != nil {
		fmt.Printf("getOrAttachBlockPeer error: %v\n", err)
		return
	}

	sess := bp.sess

	sess.mu.Lock()
	_, err = bp.ss.ReceiveMessage(payload)
	if err == nil {
		_ = os.WriteFile(sess.hotPath, sess.doc.Save(), 0o644)
	}
	sess.mu.Unlock()
	if err != nil {
		fmt.Printf("ReceiveMessage error: %v\n", err)
		return
	}

	sh.broadcastSyncBlock(sess, blockID, c)

	changeMsg := []byte(fmt.Sprintf(`{"kind":"block-changed","blockId":%q}`, blockID))
	sh.mgr.BroadcastNotebookText(c.nbID, c, changeMsg)
}

func (sh *SocketHandler) handleText(c *Client, data []byte) {
	var msg map[string]any
	if err := json.Unmarshal(data, &msg); err != nil {
		return
	}
	kind, _ := msg["kind"].(string)

	switch kind {
	case "presence":
		// msg["blockId"] и т.д.
		sh.mgr.BroadcastNotebookText(c.nbID, c, data)

	case "block-select":
		// UX (где сейчас курсор/фокус у пользователя)
		sh.mgr.BroadcastNotebookText(c.nbID, c, data)

	case "block-join":
		// пользователь открыл блок
		blockID, _ := msg["blockId"].(string)
		_, _ = sh.getOrAttachBlockPeer(c, blockID)

	case "block-done":
		blockID, _ := msg["blockId"].(string)
		sh.mgr.Save(context.Background(), c, blockID)

	default:
		// любые другие текстовые сообщения по ноутбуку
		sh.mgr.BroadcastNotebookText(c.nbID, c, data)
	}
}

func (sh *SocketHandler) broadcastSyncBlock(sess *DocSession, blockID string, from *Client) {
	sess.mu.Lock()
	clients := make([]*Client, 0, len(sess.clients))
	for c := range sess.clients {
		clients = append(clients, c)
	}
	sess.mu.Unlock()

	for _, c := range clients {
		// if from != nil && c == from {
		// 	continue
		// }
		bp := c.blocks[blockID]
		if bp == nil {
			continue
		}

		sess.mu.Lock()
		for {
			sm, ok := bp.ss.GenerateMessage()
			if !ok {
				break
			}
			frame := makeFrame(blockID, sm.Bytes())
			select {
			case c.send <- OutMsg{Type: websocket.BinaryMessage, Data: frame}:
			default:
				// медленный клиент — по ситуации: дроп/закрыть
			}
		}
		sess.mu.Unlock()
	}
}

func (sh *SocketHandler) getOrAttachBlockPeer(c *Client, blockID string) (*BlockPeer, error) {
	if bp := c.blocks[blockID]; bp != nil {
		return bp, nil
	}

	// новый блок для этого клиента
	docID := c.nbID + "::" + blockID

	sess, err := sh.mgr.Acquire(context.Background(), docID)
	if err != nil {
		return nil, err
	}

	// подписываем клиента на этот DocSession
	sess.mu.Lock()
	if sess.clients == nil {
		sess.clients = make(map[*Client]struct{})
	}
	sess.clients[c] = struct{}{}
	sess.mu.Unlock()

	bp := &BlockPeer{
		sess: sess,
		ss:   automerge.NewSyncState(sess.doc),
	}
	c.blocks[blockID] = bp

	return bp, nil
}

func (sh *SocketHandler) RegisterRoutes(r *router.Group) {
	r.GET("/ws", middleware.CommonMW(sh.Handle))
}
