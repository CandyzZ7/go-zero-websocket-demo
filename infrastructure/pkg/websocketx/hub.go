package websocketx

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/zeromicro/go-zero/core/logx"

	"go-zero-websocket-demo/infrastructure/pkg/rediskey"
	"go-zero-websocket-demo/internal/repository"
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

type HubInfo struct {
	ClientsLen        int      `json:"clientsLen"`        // 客户端连接数
	UsersLen          int      `json:"usersLen"`          // 登录用户数
	ChanRegisterLen   int      `json:"chanRegisterLen"`   // 未处理连接事件数
	ChanLoginLen      int      `json:"chanLoginLen"`      // 未处理登录事件数
	ChanUnregisterLen int      `json:"chanUnregisterLen"` // 未处理退出登录事件数
	ChanBroadcastLen  int      `json:"chanBroadcastLen"`  // 未处理广播事件数
	ClientAddrList    []string `json:"clientAddrList"`    // 客户端列表
	UserList          []string `json:"userList"`          // 登录用户列表
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
			// 广播消息
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
	h.DelClientList(client)

	// 删除用户连接
	deleteResult := h.DelUserList(client)
	if deleteResult == false {
		// 不是当前连接的客户端
		return
	}

	// 获取用户登录信息
	userOnlineEntity, err := repository.GetUserOnlineByAppIDAndUserID(context.Background(), client.AppID, client.UserID)
	if err != nil {
		logx.Error(err)
	}

	// 更新用户在线状态
	userOnlineEntity.Logout()

	// 更新用户在线状态
	err = repository.UpdateUserOnline(context.Background(), userOnlineEntity)
	if err != nil {
		logx.Error(err)
	}

	logx.Infof("client disconnect, addr: %s, appID: %s, userID: %s", client.Addr, client.AppID, client.UserID)
}

// EventRegister 用户建立连接事件
func (h *Hub) EventRegister(client *Client) {
	h.AddClientList(client)
	logx.Infof("client register, addr: %s, appID: %s, userID: %s", client.Addr, client.AppID, client.UserID)
}

// AddClientList 添加客户端
func (h *Hub) AddClientList(client *Client) {
	h.ClientsLock.Lock()
	defer h.ClientsLock.Unlock()
	h.Clients[client] = true
}

// DelClientList 删除客户端
func (h *Hub) DelClientList(client *Client) {
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

// DelUserList 删除用户
func (h *Hub) DelUserList(client *Client) (result bool) {
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

// GetClientList 获取所有客户端
func (h *Hub) GetClientList() (clients map[*Client]bool) {
	clients = make(map[*Client]bool)
	h.ClientListRange(func(client *Client, value bool) (result bool) {
		clients[client] = value
		return true
	})
	return
}

// ClientListRange 遍历
func (h *Hub) ClientListRange(f func(client *Client, value bool) (result bool)) {
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

// GetClientAddrList 获取客户端地址列表
func (h *Hub) GetClientAddrList() (clientAddrList []string) {
	clientAddrList = make([]string, 0)
	h.ClientListRange(func(client *Client, value bool) (result bool) {
		clientAddrList = append(clientAddrList, client.Addr)
		return true
	})
	return
}

// GetAllClientList 获取客户端列表
func (h *Hub) GetAllClientList() (clientList []*Client) {
	clientList = make([]*Client, 0)
	h.ClientListRange(func(client *Client, value bool) (result bool) {
		clientList = append(clientList, client)
		return true
	})
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

// GetUserKeyList 获取用户的key
func (h *Hub) GetUserKeyList() (userKeys []string) {
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

// GetUserClientList 获取用户的key
func (h *Hub) GetUserClientList() (clients []*Client) {
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
	clients := h.GetUserClientList()
	for _, conn := range clients {
		if conn != ignoreClient {
			conn.SendMsg(message)
		}
	}
}

// sendAppIDAll 向全部成员(除了自己)发送数据
func (h *Hub) sendAppIDAll(message []byte, appID string, ignoreClient *Client) {
	clients := h.GetUserClientList()
	for _, conn := range clients {
		if conn != ignoreClient && conn.AppID == appID {
			conn.SendMsg(message)
		}
	}
}

// GetHubInfo 获取管理者信息
func (h *Hub) GetHubInfo() *HubInfo {
	clientsLen := h.GetClientsLen()        // 客户端连接数
	usersLen := h.GetUsersLen()            // 登录用户数
	chanRegisterLen := len(h.Register)     // 未处理连接事件数
	chanLoginLen := len(h.Login)           // 未处理登录事件数
	chanUnregisterLen := len(h.Unregister) // 未处理退出登录事件数
	chanBroadcastLen := len(h.Broadcast)   // 未处理广播事件数
	clientAddrList := make([]string, 0)
	h.ClientListRange(func(client *Client, value bool) (result bool) {
		clientAddrList = append(clientAddrList, client.Addr)
		return true
	})
	userList := h.GetUserKeyList()

	return &HubInfo{
		ClientsLen:        clientsLen,
		UsersLen:          usersLen,
		ChanRegisterLen:   chanRegisterLen,
		ChanLoginLen:      chanLoginLen,
		ChanUnregisterLen: chanUnregisterLen,
		ChanBroadcastLen:  chanBroadcastLen,
		ClientAddrList:    clientAddrList,
		UserList:          userList,
	}
}

// ClearTimeoutConnections 定时清理超时连接
func (h *Hub) ClearTimeoutConnections() {
	currentTime := uint64(time.Now().Unix())
	clients := h.GetClientList()
	for client := range clients {
		if client.IsHeartbeatTimeout(currentTime) {
			logx.Infof("heartbeat timeout, close connection, addr: %s, appID: %s,  loginTime: %d, heartbeatTime: %d", client.Addr, client.UserID, client.LoginTime, client.HeartbeatTime)
			_ = client.Conn.Close()
		}
	}
}

// AllSendMessages 全员广播
func (h *Hub) AllSendMessages(appID string, userID string, data string) {
	logx.Infof("all send message, appID: %s, userID: %s, data: %s", appID, userID, data)
	ignoreClient := h.GetUserClient(appID, userID)
	h.sendAppIDAll([]byte(data), appID, ignoreClient)
}
