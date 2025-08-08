package handlers

import (
	"fmt"

	"github.com/fasthttp/router"
	"github.com/fasthttp/websocket"
	"github.com/valyala/fasthttp"
)

type SocketService interface {
	Compile() error
	Run() (string, error)
	Update() error
}

var upgrader = websocket.FastHTTPUpgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(ctx *fasthttp.RequestCtx) bool { return true },
}

type SocketHandler struct {
	socketService SocketService
	subscribers map[string][]*websocket.Conn
}

func NewSocketHandler(socketService SocketService) *SocketHandler {
	return &SocketHandler{
		socketService: socketService,
		subscribers: make(map[string][]*websocket.Conn),
	}
}

func (sh *SocketHandler) Handle(ctx *fasthttp.RequestCtx) {
	err := upgrader.Upgrade(ctx, func(conn *websocket.Conn) {
		// connMesage := ConnMessage{What: "conn", Who: "domnakolesax", Offset: 0}
		// connMsg, _ := json.Marshal(connMesage)
		// for _, sub := range(subscribers["uzbek"]) {
		// 	_ = sub.WriteMessage(1, connMsg)
		// }
		// subscribers["uzbek"] = append(subscribers["uzbek"], conn)
		
		for {
			messageType, p, err := conn.ReadMessage()

			if err != nil {
				break
			}

			switch messageType {
			case -1:
				break
			case 0:
				sh.Compile(conn, p)
			case 1:
				sh.Run(conn, p)
			case 2:
				sh.Write(conn,p)
			}

			fmt.Println(p)
		}
	})

	if err != nil {
		ctx.Response.SetStatusCode(fasthttp.StatusBadRequest)
	}
}

func (sh *SocketHandler) Compile(conn *websocket.Conn, payload []byte) {}
func (sh *SocketHandler) Run(conn *websocket.Conn, payload []byte) {}
func (sh *SocketHandler) Write(conn *websocket.Conn, payload []byte) {}
func (sh *SocketHandler) RegisterRoutes(r *router.Router) {
	group := r.Group("/sockets")
	group.ANY("/", sh.Handle)
}