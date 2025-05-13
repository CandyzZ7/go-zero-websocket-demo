package websocketx

import (
	"context"
	"encoding/json"
	"github.com/zeromicro/go-zero/core/logx"
	"go-zero-websocket-demo/internal/repository"
	"go-zero-websocket-demo/pkg/rediskey"
	"sync"
	"time"
)

type Hub struct {
	Clients     map[*Client]bool   // 全部的连接
	ClientsLock sync.RWMutex       // 读写锁
	Users       map[string]*Client // 登录的用户 // appID+uuid
	UserLock    sync.RWMutex       // 读写锁
	Broadcast   chan []byte
	Register    chan *Client
	Unregister  chan *Client
	Login       chan *Client // 用户登录处理
}

func NewHub() *Hub {
	return &Hub{
		Clients:    make(map[*Client]bool),
		Users:      make(map[string]*Client),
		Register:   make(chan *Client, 1000),
		Login:      make(chan *Client, 1000),
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

// EventLogin 用户登录
func (h *Hub) EventLogin(client *Client) {
	// 连接存在，在添加
	if h.InClient(client) {
		userKey := rediskey.RedisKey(rediskey.WebSocketKey.WithParams(client.AppID)).WithSymbol(client.UserID)
		h.AddUsers(userKey, client)
	}
	logx.Infof("client login, addr: %s, appID: %s, userID: %s", client.Addr, client.AppID, client.UserID)
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
	userOnlineEntity, err := repository.GetUserOnlineByAppIDAndUserID(context.Background(), client.AppID, client.UserID)
	if err != nil {
		logx.Error(err)
	}
	userOnlineEntity.LogOutTime = uint64(time.Now().Unix())
	userOnlineEntity.IsLogoff = true
	err = repository.UpdateUserOnline(context.Background(), userOnlineEntity)
	if err != nil {
		logx.Error(err)
	}
	logx.Infof("client disconnect, addr: %s, appID: %s, userID: %s", client.Addr, client.AppID, client.UserID)
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
	userKey := rediskey.RedisKey(rediskey.WebSocketKey.WithParams(client.AppID)).WithSymbol(client.UserID)
	if value, ok := h.Users[userKey]; ok {
		// 判断是否为相同的用户
		if value.Addr != client.Addr {
			return
		}
		delete(h.Users, userKey)
		result = true
	}
	return
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
func (h *Hub) GetUserClient(appID string, userID string) (client *Client) {
	h.UserLock.RLock()
	defer h.UserLock.RUnlock()
	userKey := rediskey.RedisKey(rediskey.WebSocketKey.WithParams(appID)).WithSymbol(userID)
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
func (h *Hub) GetUserList(appID string) (userList []string) {
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
func (h *Hub) sendAppIDAll(message []byte, appID string, ignoreClient *Client) {
	clients := h.GetUserClients()
	for _, conn := range clients {
		if conn != ignoreClient && conn.AppID == appID {
			conn.SendMsg(message)
		}
	}
}
