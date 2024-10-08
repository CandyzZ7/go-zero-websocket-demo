package test

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"go-zero-websocket-demo/internal/svc"
	"go-zero-websocket-demo/internal/types"
)

type PingLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// ping
func NewPingLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PingLogic {
	return &PingLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PingLogic) Ping(req *types.PingReq) (resp *types.PingResp, err error) {
	return &types.PingResp{
		Msg: "pong",
	}, nil
}
