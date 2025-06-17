package websocketx

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"testing"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/gorilla/websocket"

	"go-zero-websocket-demo/internal/pb"
)

// WebSocketClient WebSocket 客户端结构
type WebSocketClient struct {
	conn    *websocket.Conn
	url     string
	timeout time.Duration
	done    chan struct{}
}

// NewWebSocketClient 创建新的 WebSocket 客户端
func NewWebSocketClient(url string, timeout time.Duration) *WebSocketClient {
	return &WebSocketClient{
		url:     url,
		timeout: timeout,
		done:    make(chan struct{}),
	}
}

// Connect 连接到 WebSocket 服务器
func (c *WebSocketClient) Connect() error {
	var err error
	c.conn, _, err = websocket.DefaultDialer.Dial(c.url, nil)
	if err != nil {
		return fmt.Errorf("connect error: %v", err)
	}
	log.Printf("Connected to %s", c.url)
	return nil
}

// Send 发送消息到服务器
func (c *WebSocketClient) Send(message []byte) error {
	if c.conn == nil {
		return fmt.Errorf("connection not established")
	}

	// 设置写超时
	c.conn.SetWriteDeadline(time.Now().Add(c.timeout))
	return c.conn.WriteMessage(websocket.TextMessage, message)
}

// Receive 接收服务器消息
func (c *WebSocketClient) Receive() (<-chan []byte, <-chan error) {
	msgCh := make(chan []byte)
	errCh := make(chan error, 1)

	go func() {
		defer close(msgCh)
		defer close(errCh)

		for {
			select {
			case <-c.done:
				return
			default:
				// 设置读超时
				c.conn.SetReadDeadline(time.Now().Add(c.timeout))
				_, message, err := c.conn.ReadMessage()
				if err != nil {
					errCh <- fmt.Errorf("read error: %v", err)
					return
				}
				msgCh <- message
			}
		}
	}()

	return msgCh, errCh
}

// Close 关闭 WebSocket 连接
func (c *WebSocketClient) Close() error {
	if c.conn == nil {
		return nil
	}

	close(c.done)

	// 优雅关闭连接
	err := c.conn.WriteMessage(websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		log.Printf("Error sending close message: %v", err)
	}

	// 给服务器一些时间响应
	time.Sleep(100 * time.Millisecond)

	return c.conn.Close()
}

func TestWebSocketClient(t *testing.T) {
	// 命令行参数
	serverURL := flag.String("url", "ws://localhost:8888/ws", "WebSocket server URL")
	timeout := flag.Duration("timeout", 30*time.Second, "I/O timeout")
	flag.Parse()

	// 创建并连接客户端
	client := NewWebSocketClient(*serverURL, *timeout)
	if err := client.Connect(); err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer client.Close()

	// 接收消息协程
	msgCh, errCh := client.Receive()

	// 处理接收到的消息
	go func() {
		for {
			select {
			case msg, ok := <-msgCh:
				if !ok {
					return
				}
				response := &pb.MessageResponse{}
				err := proto.Unmarshal(msg, response)
				if err != nil {
					log.Printf("Unmarshal error: %v", err)
				}
				pingResp := &pb.PingResp{}
				err = proto.Unmarshal(response.Response.Data, pingResp)
				if err != nil {
					log.Printf("Unmarshal error: %v", err)
				}
				log.Printf("Received: %v", pingResp)
			case err, ok := <-errCh:
				if !ok {
					return
				}
				log.Printf("Error: %v", err)
				return
			}
		}
	}()

	// 发送测试消息
	go func() {
		req := &pb.PingReq{Msg: "test"}
		bytes, err := proto.Marshal(req)
		if err != nil {
			log.Printf("Marshal error: %v", err)
		}
		request := &pb.MessageRequest{
			Seq:  "1111",
			Cmd:  "test.ping",
			Data: bytes,
		}
		marshal, err := proto.Marshal(request)
		if err != nil {
			log.Printf("Marshal error: %v", err)
		}
		if err := client.Send(marshal); err != nil {
			log.Printf("Send error: %v", err)
			return
		}
		log.Printf("Sent: %s", string(marshal))
		time.Sleep(2 * time.Second)
	}()

	// 等待中断信号优雅退出
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	<-interrupt
	log.Println("Interrupted, closing connection...")
}
