package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/service"

	"go-zero-websocket-demo/infrastructure/pkg/websocketx"
	"go-zero-websocket-demo/internal/svc"
)

func WsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithoutCancel(r.Context())
		conn, err := websocketx.Upgrader.Upgrade(w, r, nil)
		if err != nil {
			logc.Errorf(ctx, "error upgrading to WebSocket: %v", err)
			return
		}

		currentTime := time.Now().UnixMilli()
		h := websocketx.GetHub()
		msgType := svcCtx.Config.MsgType
		if svcCtx.Config.Mode == service.DevMode || svcCtx.Config.Mode == service.TestMode {
			if r.Header.Get("X-Content-MsgType") != "" {
				msgType = r.Header.Get("X-Content-MsgType")
			}
		}

		c := websocketx.NewClient(conn.RemoteAddr().String(), conn, currentTime, msgType)
		h.Register <- c

		go c.WritePump(ctx)
		go c.ReadPump(ctx, svcCtx)

	}
}
