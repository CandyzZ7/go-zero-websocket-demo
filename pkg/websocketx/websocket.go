package websocketx

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/zeromicro/go-zero/core/logx"
	"net/http"
	"runtime/debug"
	"sync"
	"time"
)

const (
	pongWait   = 60 * time.Second
	pingPeriod = (pongWait * 9) / 10
)

var Upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // 允许所有来源，您可能需要根据实际情况修改这个检查
	},
}

type Message struct {
	ToUserId   string `json:"to_user_id"`
	FromUserId string `json:"from_user_id"`
	Msg        string `json:"msg"`
	Type       string `json:"type"`
}

type Client struct {
	Conn          *websocket.Conn
	Hub           *Hub
	mu            sync.Mutex
	Addr          string      // 客户端地址
	Send          chan []byte // 待发送的数据
	AppID         uint32      // 登录的平台ID app/web/ios
	UserID        string      // 用户ID，用户登录以后才有
	FirstTime     uint64      // 首次连接事件
	HeartbeatTime uint64      // 用户上次心跳时间
	LoginTime     uint64      // 登录时间 登录以后才有
}

// NewClient 初始化
func NewClient(h *Hub, addr string, conn *websocket.Conn, firstTime uint64) (client *Client) {
	client = &Client{
		Hub:           h,
		Addr:          addr,
		Conn:          conn,
		Send:          make(chan []byte, 100),
		FirstTime:     firstTime,
		HeartbeatTime: firstTime,
	}
	return
}

// 用户登录
type login struct {
	AppID  uint32
	UserID string
	Client *Client
}

// GetKey 获取 key
func (l *login) GetKey() (key string) {
	key = GetUserKey(l.AppID, l.UserID)
	return
}

// GetUserKey 获取用户key
func GetUserKey(appID uint32, userID string) (key string) {
	key = fmt.Sprintf("%d_%s", appID, userID)
	return
}

type Hub struct {
	Clients     map[*Client]bool   // 全部的连接
	ClientsLock sync.RWMutex       // 读写锁
	Users       map[string]*Client // 登录的用户 // appID+uuid
	UserLock    sync.RWMutex       // 读写锁
	Broadcast   chan []byte
	Register    chan *Client
	Unregister  chan *Client
	Login       chan *login // 用户登录处理
}

func NewHub() *Hub {
	return &Hub{
		Clients:    make(map[*Client]bool),
		Users:      make(map[string]*Client),
		Register:   make(chan *Client, 1000),
		Login:      make(chan *login, 1000),
		Unregister: make(chan *Client, 1000),
		Broadcast:  make(chan []byte, 1000),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			// 建立连接事件
			h.EventRegister(client)
		case client := <-h.Unregister:
			// 断开连接事件
			h.EventUnregister(client)
		case l := <-h.Login:
			// 用户登录
			h.EventLogin(l)
		case message := <-h.Broadcast:
			h.SendMessage(message)
		}
	}
}

// EventRegister 用户建立连接事件
func (h *Hub) EventRegister(client *Client) {
	h.AddClients(client)
}

// AddClients 添加客户端
func (h *Hub) AddClients(client *Client) {
	h.ClientsLock.Lock()
	defer h.ClientsLock.Unlock()
	h.Clients[client] = true
}

// EventUnregister 用户断开连接
func (h *Hub) EventUnregister(client *Client) {
	h.DelClients(client)

	// 删除用户连接
	deleteResult := h.DelUsers(client)
	if deleteResult == false {
		// 不是当前连接的客户端
		return
	}

	// 清除redis登录数据

}

// DelClients 删除客户端
func (h *Hub) DelClients(client *Client) {
	h.ClientsLock.Lock()
	defer h.ClientsLock.Unlock()
	if _, ok := h.Clients[client]; ok {
		close(client.Send)
		err := client.Conn.Close()
		if err != nil {
			logx.Errorf("close client error: %v", err)
		}
		delete(h.Clients, client)
	}
}

// DelUsers 删除用户
func (h *Hub) DelUsers(client *Client) (result bool) {
	h.UserLock.Lock()
	defer h.UserLock.Unlock()
	key := GetUserKey(client.AppID, client.UserID)
	if value, ok := h.Users[key]; ok {
		// 判断是否为相同的用户
		if value.Addr != client.Addr {
			return
		}
		delete(h.Users, key)
		result = true
	}
	return
}

// EventLogin 用户登录
func (h *Hub) EventLogin(login *login) {
	client := login.Client
	// 连接存在，在添加
	if h.InClient(client) {
		userKey := login.GetKey()
		h.AddUsers(userKey, login.Client)
	}
	logx.Info("EventLogin 用户登录", client.Addr, login.AppID, login.UserID)
}

// AddUsers 添加用户
func (h *Hub) AddUsers(key string, client *Client) {
	h.UserLock.Lock()
	defer h.UserLock.Unlock()
	h.Users[key] = client
}

func (h *Hub) InClient(client *Client) (ok bool) {
	h.ClientsLock.RLock()
	defer h.ClientsLock.RUnlock()

	// 连接存在，在添加
	_, ok = h.Clients[client]
	return
}

func (h *Hub) SendMessage(message []byte) {
	msg := &Message{}
	_ = json.Unmarshal(message, msg)
	if msg.ToUserId != "" {
		// 只取当前客户端连接，向指定客户端通道中推送一条消息
		if client, ok := h.Users[msg.ToUserId]; ok {
			client.Send <- message
		}
	} else {
		for _, client := range h.Users {
			select {
			case client.Send <- message:
			default:
				close(client.Send)
			}
		}
	}
}

func (c *Client) WritePump() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("write stop", string(debug.Stack()), r)
		}
	}()
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Hub.Unregister <- c
	}()
	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				// 发送数据错误 关闭连接
				fmt.Println("Client发送数据 关闭连接", c.Addr, "ok", ok)
				return
			}
			err := c.Conn.WriteMessage(websocket.TextMessage, message)
			if err != nil {
				logx.Errorf("Error writing message: %v", err)
			}
		}
	}
}

func (c *Client) ReadPump() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("write stop", string(debug.Stack()), r)
		}
	}()
	defer func() {
		fmt.Println("读取客户端数据 关闭send", c)
		close(c.Send)
	}()
	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logx.Errorf("Error reading message: %v", err)
			}
			break
		}
		c.Send <- message
	}
}
