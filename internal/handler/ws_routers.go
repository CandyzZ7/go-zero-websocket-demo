package handler

import (
	"context"

	websocketx2 "go-zero-websocket-demo/infrastructure/pkg/websocketx"
	"go-zero-websocket-demo/internal/logic"
	"go-zero-websocket-demo/internal/svc"
)

func WebsocketInit(ctx context.Context, svcCtx *svc.ServiceContext, client *websocketx2.Client) {
	// 创建账户相关路由组
	accountGroup := websocketx2.NewGroup("test")

	// 注册账户相关路由
	accountGroup.Register("ping", logic.NewPingLogic(ctx, svcCtx, client).Ping)
}
