package svc

import (
	"go-zero-websocket/internal/config"
	websocket "go-zero-websocket/pkg"
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
