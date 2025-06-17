package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logx"

	"go-zero-websocket-demo/infrastructure/e"
	"go-zero-websocket-demo/infrastructure/pkg/serializex"
	"go-zero-websocket-demo/infrastructure/pkg/websocketx"
	"go-zero-websocket-demo/internal/pb"
	"go-zero-websocket-demo/internal/svc"
)

type HeartbeatLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	client *websocketx.Client
}

// Heartbeat
func NewHeartbeatLogic(ctx context.Context, svcCtx *svc.ServiceContext, client *websocketx.Client) *HeartbeatLogic {
	return &HeartbeatLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		client: client,
	}
}

func (l *HeartbeatLogic) Heartbeat(seq string, message []byte) (data []byte, err error) {
	req := &pb.HeartbeatReq{}
	err = serializex.Unmarshal(l.client.MsgType, message, req)
	if err != nil {
		return nil, err
	}

	resp, err := l.heartbeat(req)
	if err != nil {
		return nil, err
	}

	return serializex.Marshal(l.client.MsgType, resp)
}

func (l *HeartbeatLogic) heartbeat(req *pb.HeartbeatReq) (*pb.HeartbeatResp, error) {
	if l.client.IsLogin() {
		return nil, e.NotLoggedIn
	}

	return &pb.HeartbeatResp{}, nil
}
