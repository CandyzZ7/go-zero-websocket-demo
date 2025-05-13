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
	logx.Info("user login", login.AppID, login.UserID, client.Addr)
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

// GetClients 获取所有客户端
func (h *Hub) GetClients() (clients map[*Client]bool) {
	clients = make(map[*Client]bool)
	h.ClientsRange(func(client *Client, value bool) (result bool) {
		clients[client] = value
		return true
	})
	return
}

// ClientsRange 遍历
func (h *Hub) ClientsRange(f func(client *Client, value bool) (result bool)) {
	h.ClientsLock.RLock()
	defer h.ClientsLock.RUnlock()
	for key, value := range h.Clients {
		result := f(key, value)
		if result == false {
			return
		}
	}
	return
}

// GetClientsLen GetClientsLen
func (h *Hub) GetClientsLen() (clientsLen int) {
	clientsLen = len(h.Clients)
	return
}

// GetUserClient 获取用户的连接
func (h *Hub) GetUserClient(appID uint32, userID string) (client *Client) {
	h.UserLock.RLock()
	defer h.UserLock.RUnlock()
	userKey := GetUserKey(appID, userID)
	if value, ok := h.Users[userKey]; ok {
		client = value
	}
	return
}

// GetUsersLen GetClientsLen
func (h *Hub) GetUsersLen() (userLen int) {
	userLen = len(h.Users)
	return
}

// GetUserKeys 获取用户的key
func (h *Hub) GetUserKeys() (userKeys []string) {
	userKeys = make([]string, 0, len(h.Users))
	for key := range h.Users {
		userKeys = append(userKeys, key)
	}
	return
}

// GetUserList 获取用户 list
func (h *Hub) GetUserList(appID uint32) (userList []string) {
	userList = make([]string, 0)
	h.UserLock.RLock()
	defer h.UserLock.RUnlock()
	for _, v := range h.Users {
		if v.AppID == appID {
			userList = append(userList, v.UserID)
		}
	}
	return
}

// GetUserClients 获取用户的key
func (h *Hub) GetUserClients() (clients []*Client) {
	clients = make([]*Client, 0)
	h.UserLock.RLock()
	defer h.UserLock.RUnlock()
	for _, v := range h.Users {
		clients = append(clients, v)
	}
	return
}

// sendAll 向全部成员(除了自己)发送数据
func (h *Hub) sendAll(message []byte, ignoreClient *Client) {
	clients := h.GetUserClients()
	for _, conn := range clients {
		if conn != ignoreClient {
			conn.SendMsg(message)
		}
	}
}

// sendAppIDAll 向全部成员(除了自己)发送数据
func (h *Hub) sendAppIDAll(message []byte, appID uint32, ignoreClient *Client) {
	clients := h.GetUserClients()
	for _, conn := range clients {
		if conn != ignoreClient && conn.AppID == appID {
			conn.SendMsg(message)
		}
	}
}

func (c *Client) WritePump() {
	defer func() {
		if r := recover(); r != nil {
			logx.Info("write stop", string(debug.Stack()), r)
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
			err := c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok || err != nil {
				_ = c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			_, _ = w.Write(message)

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			if err := c.Conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				return
			}
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) ReadPump() {
	defer func() {
		if r := recover(); r != nil {
			logx.Info("write stop", string(debug.Stack()), r)
		}
	}()
	defer func() {
		logx.Info("client %s disconnected", c.Addr)
		close(c.Send)
	}()
	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				logx.Infof("websocket was closed unexpectedly: %v", err)
			}
			break
		}
		c.Send <- message
	}
}

// SendMsg 发送数据
func (c *Client) SendMsg(msg []byte) {
	if c == nil {
		return
	}
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("sendMsg stop:", r, string(debug.Stack()))
		}
	}()
	c.Send <- msg
}

// close 关闭客户端连接
func (c *Client) close() {
	close(c.Send)
}

// Login 用户登录
func (c *Client) Login(appID uint32, userID string, loginTime uint64) {
	c.AppID = appID
	c.UserID = userID
	c.LoginTime = loginTime
	// 登录成功=心跳一次
	c.Heartbeat(loginTime)
}

// Heartbeat 用户心跳
func (c *Client) Heartbeat(currentTime uint64) {
	c.HeartbeatTime = currentTime

	return
}

// IsHeartbeatTimeout 心跳超时
func (c *Client) IsHeartbeatTimeout(currentTime uint64) (timeout bool) {
	if c.HeartbeatTime+heartbeatExpirationTime <= currentTime {
		timeout = true
	}
	return
}

// IsLogin 是否登录了
func (c *Client) IsLogin() (isLogin bool) {
	// 用户登录了
	if c.UserID != "" {
		isLogin = true
		return
	}
	return
}
