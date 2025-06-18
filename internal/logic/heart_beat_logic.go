package logic

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"

	"go-zero-websocket-demo/infrastructure/e"
	"go-zero-websocket-demo/infrastructure/pkg/serializex"
	"go-zero-websocket-demo/infrastructure/pkg/websocketx"
	"go-zero-websocket-demo/internal/pb"
	"go-zero-websocket-demo/internal/repository"
	"go-zero-websocket-demo/internal/svc"
)

type HeartbeatLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	Client *websocketx.Client
}

// ping
func NewHeartbeatLogic(ctx context.Context, svcCtx *svc.ServiceContext, client *websocketx.Client) *HeartbeatLogic {
	return &HeartbeatLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		Client: client,
	}
}

func Heartbeat(ctx context.Context, svcCtx *svc.ServiceContext, client *websocketx.Client, seq string, message []byte) (data []byte, err error) {
	req := &pb.HeartbeatReq{}
	err = serializex.Unmarshal(client.MsgType, message, req)
	if err != nil {
		return nil, err
	}
	logc.Infof(ctx, "Received ping request: %v", req)
	resp, err := NewHeartbeatLogic(ctx, svcCtx, client).Heartbeat(req)
	if err != nil {
		return nil, err
	}
	logc.Infof(ctx, "Sending ping response: %v", resp)
	return serializex.Marshal(client.MsgType, resp)
}

func (l *HeartbeatLogic) Heartbeat(req *pb.HeartbeatReq) (*pb.HeartbeatResp, error) {
	currentTime := time.Now().UnixMilli()
	if !l.Client.IsLogin() {
		return nil, e.NotLoggedIn
	}
	userOnline, err := repository.GetUserOnlineByAppIDAndUserID(l.ctx, l.Client.AppID, l.Client.UserID)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, e.NotLoggedIn
		}
		return nil, err
	}
	l.Client.Heartbeat(currentTime)
	userOnline.Heartbeat(currentTime)
	err = repository.UpdateUserOnline(l.ctx, userOnline)
	if err != nil {
		return nil, err
	}
	return &pb.HeartbeatResp{}, nil
}
