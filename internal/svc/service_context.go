package svc

import (
	"go-zero-websocket-demo/internal/config"
	websocket "go-zero-websocket-demo/pkg"
)

type ServiceContext struct {
	Config config.Config
	WSHub  *websocket.Hub
	// 其他服务...
}

func NewServiceContext(c config.Config) *ServiceContext {
	wsHub := websocket.NewHub()
	go wsHub.Run()

	return &ServiceContext{
		Config: c,
		WSHub:  wsHub,
		// 初始化其他服务...
	}
}
