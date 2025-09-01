package websocketx

import (
	"context"
	"fmt"
	"io"
	"runtime/debug"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/zeromicro/go-zero/core/logx"

	"go-zero-websocket-demo/common/rediskey"
	"go-zero-websocket-demo/internal/svc"
)

type Client struct {
	Conn          *websocket.Conn
	mu            sync.Mutex
	Addr          string      // 客户端地址
	Send          chan []byte // 待发送的数据
	AppID         string      // 登录的平台ID app/web/ios
	UserID        string      // 用户ID，用户登录以后才有
	FirstTime     int64       // 首次连接事件
	HeartbeatTime int64       // 用户上次心跳时间
	LoginTime     int64       // 登录时间 登录以后才有
	MsgType       string
}

// NewClient 初始化
func NewClient(addr string, conn *websocket.Conn, firstTime int64, msgType string) (client *Client) {
	client = &Client{
		Addr:          addr,
		Conn:          conn,
		Send:          make(chan []byte, 100),
		FirstTime:     firstTime,
		HeartbeatTime: firstTime,
		MsgType:       msgType,
	}
	return
}

func (c *Client) WritePump(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			logx.Info("write stop", string(debug.Stack()), r)
		}
	}()
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		GetHub().Unregister <- c
	}()
	for {
		select {
		case message, ok := <-c.Send:
			err := c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok || err != nil {
				_ = c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			var w io.WriteCloser
			if c.MsgType == "proto" {
				w, err = c.Conn.NextWriter(websocket.BinaryMessage)
				if err != nil {
					return
				}
			}
			if c.MsgType == "json" {
				w, err = c.Conn.NextWriter(websocket.TextMessage)
				if err != nil {
					return
				}
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

func (c *Client) ReadPump(ctx context.Context, svcCtx *svc.ServiceContext) {
	defer func() {
		if r := recover(); r != nil {
			logx.Info("write stop", string(debug.Stack()), r)
		}
	}()
	defer func() {
		logx.Infof("client %s disconnected", c.Addr)
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
		ProcessData(ctx, svcCtx, c, message)
	}
}

func (c *Client) GetUserKey() string {
	return rediskey.RedisKey(rediskey.WebSocketKey.WithParams(c.AppID)).WithSymbol(c.UserID)
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
func (c *Client) Login(appID string, userID string, loginTime int64) {
	c.AppID = appID
	c.UserID = userID
	c.LoginTime = loginTime
	// 登录成功=心跳一次
	c.Heartbeat(loginTime)
}

// Heartbeat 用户心跳
func (c *Client) Heartbeat(currentTime int64) {
	c.HeartbeatTime = currentTime

	return
}

// IsHeartbeatTimeout 心跳超时
func (c *Client) IsHeartbeatTimeout(currentTime int64) (timeout bool) {
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
