package logic

import (
	"context"

	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"

	"go-zero-websocket-demo/infrastructure/pkg/serializex"
	"go-zero-websocket-demo/infrastructure/pkg/websocketx"
	"go-zero-websocket-demo/internal/pb"
	"go-zero-websocket-demo/internal/svc"
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

func (l *PingLogic) Ping(client *websocketx.Client, seq string, message []byte) (data []byte, err error) {
	req := &pb.PingReq{}
	err = serializex.Unmarshal(client.MsgType, message, req)
	if err != nil {
		return nil, err
	}
	logc.Infof(l.ctx, "Received ping request: %v", req)
	resp, err := l.ping(req)
	if err != nil {
		return nil, err
	}
	logc.Infof(l.ctx, "Sending ping response: %v", resp)
	return serializex.Marshal(client.MsgType, resp)
}

func (l *PingLogic) ping(req *pb.PingReq) (*pb.PingResp, error) {
	return &pb.PingResp{
		Msg: "test" + req.Msg,
	}, nil
}
