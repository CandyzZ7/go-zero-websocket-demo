package logic

import (
	"context"
	"encoding/json"

	"github.com/zeromicro/go-zero/core/logx"

	"go-zero-websocket-demo/infrastructure/pkg/websocketx"
	"go-zero-websocket-demo/internal/pb"
	"go-zero-websocket-demo/internal/svc"
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

func (l *PingLogic) Ping(seq string, message []byte) (data []byte, err error) {
	var req *pb.PingReq
	err = json.Unmarshal(message, &req)
	if err != nil {
		return nil, err
	}
	resp, err := l.ping(req)
	if err != nil {
		return nil, err
	}
	// 将响应序列化为JSON
	respBytes, err := json.Marshal(resp)
	if err != nil {
		return nil, err
	}
	return respBytes, nil
}

func (l *PingLogic) ping(req *pb.PingReq) (*pb.PingResp, error) {
	return &pb.PingResp{
		Msg: "test" + req.Msg,
	}, nil
}
