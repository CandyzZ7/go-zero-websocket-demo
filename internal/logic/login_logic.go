package logic

import (
	"context"
	"time"

	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"

	"go-zero-websocket-demo/common/ctxdata"
	"go-zero-websocket-demo/common/rediskey"
	"go-zero-websocket-demo/infrastructure/e"
	"go-zero-websocket-demo/infrastructure/pkg/serializex"
	"go-zero-websocket-demo/infrastructure/pkg/websocketx"
	"go-zero-websocket-demo/infrastructure/pkg/xcrypt"
	"go-zero-websocket-demo/infrastructure/pkg/xjwt"
	"go-zero-websocket-demo/internal/model"
	"go-zero-websocket-demo/internal/pb"
	"go-zero-websocket-demo/internal/repository"
	"go-zero-websocket-demo/internal/svc"
)

type LoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	Client *websocketx.Client
}

// ping
func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext, client *websocketx.Client) *LoginLogic {
	return &LoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		Client: client,
	}
}

func Login(ctx context.Context, svcCtx *svc.ServiceContext, client *websocketx.Client, seq string, message []byte) (data []byte, err error) {
	req := &pb.LoginReq{}
	err = serializex.Unmarshal(client.MsgType, message, req)
	if err != nil {
		return nil, err
	}
	logc.Infof(ctx, "Received ping request: %v", req)
	resp, err := NewLoginLogic(ctx, svcCtx, client).Login(req)
	if err != nil {
		return nil, err
	}
	logc.Infof(ctx, "Sending ping response: %v", resp)
	return serializex.Marshal(client.MsgType, resp)
}

func (l *LoginLogic) Login(req *pb.LoginReq) (*pb.LoginResp, error) {
	currentTime := time.Now().UnixMilli()
	// 验证 Token 是否有效
	claims, err := xjwt.ValidateToken(req.Token, l.svcCtx.Config.TokenConf.AccessSecret)
	if err != nil {
		return nil, err
	}
	// 验证 Token 是否存在
	tokenMd5 := xcrypt.EncryptMD5(req.Token)
	existed, err := l.svcCtx.RDB.ExistsCtx(l.ctx, rediskey.TokenKey.WithParams(tokenMd5))
	if err != nil {
		return nil, err
	}
	if !existed {
		return nil, e.ErrTokenVerifyFail
	}
	userId, ok := claims[ctxdata.CtxKeyJwtUserId].(string)
	if !ok {
		return nil, e.ErrTokenVerifyFail
	}
	appId, ok := claims[ctxdata.CtxKeyJwtAPPId].(string)
	if !ok {
		return nil, e.ErrTokenVerifyFail
	}
	l.Client.Login(appId, userId, currentTime)
	userOnline := &model.UserOnline{}
	userOnline.Login(appId, userId, l.Client.Addr, currentTime)
	err = repository.UpdateUserOnline(l.ctx, userOnline)
	if err != nil {
		return nil, err
	}

	hub := websocketx.GetHub()
	hub.Login <- l.Client
	return &pb.LoginResp{}, nil
}
