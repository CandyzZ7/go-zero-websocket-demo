package handler

import (
	"net/http"
	"time"

	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/service"

	websocketx "go-zero-websocket-demo/infrastructure/pkg/websocketx"
	"go-zero-websocket-demo/internal/svc"
)

func WsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocketx.Upgrader.Upgrade(w, r, nil)
		if err != nil {
			logc.Errorf(r.Context(), "error upgrading to WebSocket: %v", err)
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

		c := websocketx.NewClient(h, conn.RemoteAddr().String(), conn, currentTime, msgType)
		h.Register <- c

		go c.WritePump(r.Context())
		go c.ReadPump(r.Context(), svcCtx)

	}
}
