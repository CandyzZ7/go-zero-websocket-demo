package test

import (
	"context"
	"go-zero-websocket-demo/pkg/websocketx"

	"github.com/zeromicro/go-zero/core/logx"

	"go-zero-websocket-demo/internal/svc"
	"go-zero-websocket-demo/internal/types"
)

type PingLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	client *websocketx.Client
}

// ping
func NewPingLogic(ctx context.Context, svcCtx *svc.ServiceContext, client *websocketx.Client) *PingLogic {
	return &PingLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		client: client,
	}
}

func (l *PingLogic) Ping(req *types.PingReq) (resp *types.PingResp, err error) {
	return &types.PingResp{
		Msg: "pong",
	}, nil
}
