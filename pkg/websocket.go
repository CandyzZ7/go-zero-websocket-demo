package pkg

import (
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/zeromicro/go-zero/core/logx"
)

var Upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // 允许所有来源，您可能需要根据实际情况修改这个检查
	},
}

type Connection struct {
	Conn *websocket.Conn
	mu   sync.Mutex
}

type Hub struct {
	Connections map[*Connection]bool
	Broadcast   chan []byte
	Register    chan *Connection
	Unregister  chan *Connection
}

func NewHub() *Hub {
	return &Hub{
		Connections: make(map[*Connection]bool),
		Broadcast:   make(chan []byte),
		Register:    make(chan *Connection),
		Unregister:  make(chan *Connection),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case conn := <-h.Register:
			h.Connections[conn] = true
		case conn := <-h.Unregister:
			if _, ok := h.Connections[conn]; ok {
				delete(h.Connections, conn)
				conn.Conn.Close()
			}
		case message := <-h.Broadcast:
			for conn := range h.Connections {
				conn.Write(message)
			}
		}
	}
}

func (c *Connection) Write(message []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	err := c.Conn.WriteMessage(websocket.TextMessage, message)
	if err != nil {
		logx.Errorf("Error writing message: %v", err)
	}
}

func (h *Hub) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := Upgrader.Upgrade(w, r, nil)
	if err != nil {
		logx.Errorf("Error upgrading to WebSocket: %v", err)
		return
	}

	c := &Connection{Conn: conn}
	h.Register <- c

	defer func() {
		h.Unregister <- c
	}()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logx.Errorf("Error reading message: %v", err)
			}
			break
		}
		h.Broadcast <- message
	}
}
