package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/fasthttp/router"
	"github.com/fasthttp/websocket"
	"github.com/valyala/fasthttp"
)

var upgrader = websocket.FastHTTPUpgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(ctx *fasthttp.RequestCtx) bool { return true },
}

type SocketHandler struct {
	mgr *Manager
}

func NewHandler(mgr *Manager) *SocketHandler {
	return &SocketHandler{mgr: mgr}
}

func (sh *SocketHandler) Handle(ctx *fasthttp.RequestCtx) {
	docID := ctx.Request.URI().QueryArgs().Peek("nb")
	if docID == nil {
		// http.Error(w, "missing doc", http.StatusBadRequest)
		fmt.Println("missing doc")
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}

	defer func() {
		res := recover()

		if res != nil {
			fmt.Printf("PANIC OCCURED: %v" ,res)
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
	err := upgrader.Upgrade(ctx, func(conn *websocket.Conn) {
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
	sess := c.sess
	if sess == nil {
		// клиент ещё не выбрал блок — игнорируем
		return
	}

	sess.mu.Lock()
	_, err := c.ss.ReceiveMessage(msg)
	if err == nil {
		// обновляем hot-снапшот
		_ = os.WriteFile(sess.hotPath, sess.doc.Save(), 0o644)
	}
	sess.mu.Unlock()
	if err != nil {
		return
	}

	// после применения sync — рассылаем изменения всем клиентам этого блока
	sh.broadcastSync(sess)

	// опционально: шлём текстовое уведомление "block-changed" по ноутбуку
	changeMsg := []byte(fmt.Sprintf(`{"kind":"block-changed","blockId":%q}`, c.blockID))
	sh.mgr.BroadcastNotebookText(c.nbID, c, changeMsg)
}

func (sh *SocketHandler) handleText(c *Client, data []byte) {
	// грубый быстрый разбор "kind"
	var msg map[string]any
	if err := json.Unmarshal(data, &msg); err != nil {
		return
	}
	kind, _ := msg["kind"].(string)

	switch kind {
	case "block-select":
		blockID, _ := msg["blockId"].(string)
		if blockID == "" {
			return
		}

		// смена блока
		sess, err := sh.mgr.SwitchBlock(context.Background(), c, blockID)
		if err != nil {
			return
		}

		// отправляем текущий state блока этому клиенту
		sh.kickSync(sess, c)

		// сообщаем всем клиентам ноутбука, что этот пользователь теперь на blockID
		sh.mgr.BroadcastNotebookText(c.nbID, c, data)

	default:
		// presence / любые другие текстовые сообщения — ретранслируем
		// по ноутбуку (чтобы все знали, что за блок затронут)
		sh.mgr.BroadcastNotebookText(c.nbID, c, data)
	}
}

func (sh *SocketHandler) kickSync(sess *DocSession, c *Client) {
	sess.mu.Lock()
	defer sess.mu.Unlock()

	for {
		sm, valid := c.ss.GenerateMessage()
		if !valid {
			break
		}
		select {
		case c.send <- OutMsg{Type: websocket.BinaryMessage, Data: sm.Bytes()}:
		default:
			return
		}
	}
}

func (sh *SocketHandler) broadcastSync(sess *DocSession) {
	sess.mu.Lock()
	clients := make([]*Client, 0, len(sess.clients))
	for c := range sess.clients {
		clients = append(clients, c)
	}
	sess.mu.Unlock()

	for _, c := range clients {
		sh.kickSync(sess, c)
	}
}

func (sh *SocketHandler) RegisterRoutes(r *router.Router) {
	r.GET("/ws", sh.Handle)
}
