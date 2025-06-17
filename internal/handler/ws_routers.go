package handler

import (
	"go-zero-websocket-demo/infrastructure/pkg/websocketx"
	"go-zero-websocket-demo/internal/logic"
)

func WebsocketInit() {
	// 创建账户相关路由组
	accountGroup := websocketx.NewGroup("test")

	// 注册账户相关路由
	accountGroup.Register("ping", logic.Ping)
}
