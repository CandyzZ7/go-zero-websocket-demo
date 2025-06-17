package websocketx

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/golang/protobuf/proto"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"

	"go-zero-websocket-demo/infrastructure/e"
	"go-zero-websocket-demo/internal/pb"
)

// Request 通用请求数据格式
type Request struct {
	Seq  string      `json:"seq"`            // 消息的唯一ID
	Cmd  string      `json:"cmd"`            // 请求命令字
	Data interface{} `json:"data,omitempty"` // 数据 json
}

// Login 登录请求数据
type Login struct {
	ServiceToken string `json:"serviceToken"` // 验证用户是否登录
	AppID        uint32 `json:"appID,omitempty"`
	UserID       string `json:"userID,omitempty"`
}

// HeartBeat 心跳请求数据
type HeartBeat struct {
	UserID string `json:"userID,omitempty"`
}

// Head 响应数据头
type Head struct {
	Seq      string    `json:"seq"`      // 消息的ID
	Cmd      string    `json:"cmd"`      // 消息的cmd 动作
	Response *Response `json:"response"` // 消息体
}

// Response 响应数据体
type Response struct {
	Code uint32      `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"` // 数据 json
}

// PushMsg 数据结构体
type PushMsg struct {
	Seq  string `json:"seq"`
	Uuid uint64 `json:"uuid"`
	Type string `json:"type"`
	Msg  string `json:"msg"`
}

// NewJsonMessageResponse 设置返回消息
func NewJsonMessageResponse(seq string, cmd string, code *e.StatusCode, data []byte) *Head {
	response := NewJsonResponse(code, data)
	return &Head{Seq: seq, Cmd: cmd, Response: response}
}

// NewJsonResponse 创建新的响应
func NewJsonResponse(code *e.StatusCode, data []byte) *Response {
	// 用于存放解析结果的map
	var result map[string]interface{}

	// 解析JSON数据
	err := json.Unmarshal(data, &result)
	if err != nil {
		logx.Errorf("json.Unmarshal error: %v", err)
	}
	return &Response{Code: uint32(code.Code), Msg: code.Message, Data: result}
}

// NewProtoMessageResponse 设置返回消息
func NewProtoMessageResponse(seq string, cmd string, code *e.StatusCode, data []byte) *pb.MessageResponse {
	response := NewProtoResponse(code, data)
	return &pb.MessageResponse{Seq: seq, Cmd: cmd, Response: response}
}

// NewProtoResponse 创建新的响应
func NewProtoResponse(code *e.StatusCode, data []byte) *pb.Response {
	return &pb.Response{Code: int32(code.Code), Msg: code.Message, Data: data}
}

// String to string
func (h *Head) String() (headStr string) {
	headBytes, _ := json.Marshal(h)
	headStr = string(headBytes)
	return
}

// DisposeFunc 处理函数（修改data类型为[]byte）
type DisposeFunc func(seq string, message []byte) (data []byte, err error)

// MiddlewareFunc 中间件函数类型（修改data类型为[]byte）
type MiddlewareFunc func(client *Client, seq string, cmd string, message []byte, next DisposeFunc) (data []byte, err error)

// Route 表示一个路由配置
type Route struct {
	Handler     DisposeFunc
	Middlewares []MiddlewareFunc
}

// Group 表示一个路由组
type Group struct {
	prefix      string
	middlewares []MiddlewareFunc
}

var (
	routes        = make(map[string]Route)
	routesRWMutex sync.RWMutex
)

// Register 注册路由及其中间件
func Register(key string, handler DisposeFunc, middlewares ...MiddlewareFunc) {
	routesRWMutex.Lock()
	defer routesRWMutex.Unlock()
	routes[key] = Route{Handler: handler, Middlewares: middlewares}
}

// Use 为路由添加中间件
func Use(key string, middleware MiddlewareFunc) {
	routesRWMutex.Lock()
	defer routesRWMutex.Unlock()
	route, ok := routes[key]
	if !ok {
		route = Route{Middlewares: []MiddlewareFunc{}}
	}
	route.Middlewares = append(route.Middlewares, middleware)
	routes[key] = route
}

// NewGroup 创建一个新的路由组
func NewGroup(prefix string, middlewares ...MiddlewareFunc) *Group {
	return &Group{prefix: prefix, middlewares: middlewares}
}

// Use 为路由组添加中间件
func (g *Group) Use(middlewares ...MiddlewareFunc) {
	g.middlewares = append(g.middlewares, middlewares...)
}

// Register 在路由组内注册路由
func (g *Group) Register(key string, handler DisposeFunc, middlewares ...MiddlewareFunc) {
	fullKey := g.prefix + "." + key
	allMiddlewares := make([]MiddlewareFunc, 0, len(g.middlewares)+len(middlewares))
	allMiddlewares = append(allMiddlewares, g.middlewares...)
	allMiddlewares = append(allMiddlewares, middlewares...)
	Register(fullKey, handler, allMiddlewares...)
}

func getRoute(key string) (route Route, ok bool) {
	routesRWMutex.RLock()
	defer routesRWMutex.RUnlock()
	route, ok = routes[key]
	return
}

// ProcessData 处理数据
func ProcessData(ctx context.Context, client *Client, message []byte, msgType string) {
	logc.Infof(ctx, "client Received message:%s,addr:%s,appID:%s,userID:%s", message, client.Addr, client.AppID, client.UserID)
	defer func() {
		if r := recover(); r != nil {
			logc.Error(ctx, "process data panic", r)
			sendErrorResponse(ctx, client, msgType, e.SystemError, "")
		}
	}()

	// 消息处理逻辑
	var (
		seq, cmd string
		data     []byte
		status   = e.OK
	)

	// 解析请求
	requestData, err := parseRequest(ctx, msgType, message)
	if err != nil {
		status = e.ParseError
		sendErrorResponse(ctx, client, msgType, status, cmd)
		return
	}
	seq, cmd, data = requestData.Seq, requestData.Cmd, requestData.Data

	// 路由处理
	route, ok := getRoute(cmd)
	if !ok {
		status = e.NotFoundRoute
		sendErrorResponse(ctx, client, msgType, status, cmd)
		logc.Error(ctx, "route not found", client.Addr, "cmd", cmd)
		return
	}

	// 创建处理链并执行
	handler := chainMiddlewares(route.Handler, route.Middlewares, client, cmd)
	responseData, err := handler(seq, data)
	if err != nil {
		status = e.ErrHandler(err)
	}

	// 构建并发送响应
	if err := sendSuccessResponse(ctx, client, msgType, seq, cmd, status, responseData); err != nil {
		logc.Error(ctx, "send response failed", err)
	}

}

// RequestData 封装解析后的请求数据
type RequestData struct {
	Seq  string
	Cmd  string
	Data []byte
}

// parseRequest 解析不同类型的请求
func parseRequest(ctx context.Context, msgType string, message []byte) (*RequestData, error) {
	var (
		seq, cmd string
		data     []byte
		err      error
	)

	switch msgType {
	case "json":
		request := &Request{}
		if err = json.Unmarshal(message, request); err != nil {
			logc.Error(ctx, "json Unmarshal ", err)
			return nil, err
		}
		if data, err = json.Marshal(request.Data); err != nil {
			logc.Error(ctx, "json Marshal ", err)
			return nil, err
		}
		seq, cmd = request.Seq, request.Cmd

	case "proto":
		request := &pb.MessageRequest{}
		if err = proto.Unmarshal(message, request); err != nil {
			logc.Error(ctx, "proto Unmarshal ", err)
			return nil, err
		}
		seq, cmd, data = request.Seq, request.Cmd, request.Data

	default:
		return nil, fmt.Errorf("unsupported message type: %s", msgType)
	}

	return &RequestData{Seq: seq, Cmd: cmd, Data: data}, nil
}

// sendSuccessResponse 发送成功响应
func sendSuccessResponse(ctx context.Context, client *Client, msgType, seq, cmd string, status *e.StatusCode, data []byte) error {
	var response interface{}
	switch msgType {
	case "json":
		response = NewJsonMessageResponse(seq, cmd, status, data)
	case "proto":
		response = NewProtoMessageResponse(seq, cmd, status, data)
	default:
		return fmt.Errorf("unsupported message type: %s", msgType)
	}

	return sendResponse(ctx, client, response, msgType)
}

// sendErrorResponse 发送错误响应
func sendErrorResponse(ctx context.Context, client *Client, msgType string, status *e.StatusCode, cmd string) {
	var response interface{}
	switch msgType {
	case "json":
		response = NewJsonMessageResponse("", cmd, status, []byte(""))
	case "proto":
		response = NewProtoMessageResponse("", cmd, status, []byte(""))
	default:
		logc.Error(ctx, "no such message type", msgType)
		return
	}

	if err := sendResponse(ctx, client, response, msgType); err != nil {
		logc.Error(ctx, "send error response failed", err)
	}
}

// sendResponse 发送响应消息 (修改为返回错误)
func sendResponse(ctx context.Context, client *Client, response interface{}, msgType string) error {
	var responseBytes []byte
	var err error

	switch msgType {
	case "json":
		responseBytes, err = json.Marshal(response)
	case "proto":
		responseBytes, err = proto.Marshal(response.(proto.Message))
	default:
		err = fmt.Errorf("unsupported message type: %s", msgType)
	}

	if err != nil {
		logc.Error(ctx, "serialize response", err)
		return err
	}

	client.SendMsg(responseBytes)
	logc.Infof(ctx, "client send message:%s, addr:%s,appID:%s,userID:%s", responseBytes, client.Addr, client.AppID, client.UserID)
	return nil
}

// chainMiddlewares 将中间件链接成处理链
func chainMiddlewares(final DisposeFunc, middlewares []MiddlewareFunc, client *Client, cmd string) DisposeFunc {
	// 直接返回最终处理函数（无中间件）
	if len(middlewares) == 0 {
		return final
	}

	// 从后向前构建中间件链（保持注册顺序）
	return func(seq string, message []byte) ([]byte, error) {
		// 当前处理函数初始化为最终处理函数
		currentHandler := final

		// 从最后一个中间件开始向前构建链
		for i := len(middlewares) - 1; i >= 0; i-- {
			middleware := middlewares[i]

			// 保存前一个处理函数
			prevHandler := currentHandler

			// 创建新的处理函数，将当前中间件和前一个处理函数链接起来
			currentHandler = func(s string, msg []byte) ([]byte, error) {
				return middleware(client, s, cmd, msg, prevHandler)
			}
		}

		// 执行构建好的中间件链
		return currentHandler(seq, message)
	}
}
