package handler

import (
	"context"

	"go-zero-websocket-demo/infrastructure/pkg/websocketx"
	"go-zero-websocket-demo/internal/logic"
	"go-zero-websocket-demo/internal/svc"
)

func WebsocketInit(ctx context.Context, svcCtx *svc.ServiceContext) {
	// 创建账户相关路由组
	accountGroup := websocketx.NewGroup("test")

	// 注册账户相关路由
	accountGroup.Register("ping", logic.NewPingLogic(ctx, svcCtx).Ping)
}
