package svc

import (
	"go-zero-websocket-demo/internal/config"
	"go-zero-websocket-demo/pkg/websocketx"
)

type ServiceContext struct {
	Config config.Config
	WSHub  *websocketx.Hub
	// 其他服务...
}

func NewServiceContext(c config.Config) *ServiceContext {
	wsHub := websocketx.NewHub()
	go wsHub.Run()

	return &ServiceContext{
		Config: c,
		WSHub:  wsHub,
		// 初始化其他服务...
	}
}
