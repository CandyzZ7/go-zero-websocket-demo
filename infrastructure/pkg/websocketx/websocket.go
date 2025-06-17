package websocketx

import (
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait               = 10 * time.Second
	pongWait                = 60 * time.Second
	pingPeriod              = (pongWait * 9) / 10
	heartbeatExpirationTime = 6 * 60
)

var Upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // 允许所有来源，您可能需要根据实际情况修改这个检查
	},
}

type Message struct {
	Seq        string `json:"seq"` // 消息的唯一ID
	Cmd        string `json:"cmd"` // 请求命令字
	ToUserId   string `json:"to_user_id"`
	FromUserId string `json:"from_user_id"`
	Msg        string `json:"msg"`
	Type       string `json:"type"`
}
