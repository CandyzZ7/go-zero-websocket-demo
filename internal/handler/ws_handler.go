package handler

import (
	"net/http"
	"time"

	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/service"

	websocketx2 "go-zero-websocket-demo/infrastructure/pkg/websocketx"
	"go-zero-websocket-demo/internal/svc"
)

func WsHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocketx2.Upgrader.Upgrade(w, r, nil)
		if err != nil {
			logc.Errorf(r.Context(), "error upgrading to WebSocket: %v", err)
			return
		}

		currentTime := uint64(time.Now().Unix())
		h := svcCtx.WSHub
		c := websocketx2.NewClient(h, conn.RemoteAddr().String(), conn, currentTime, svcCtx.Config)
		h.Register <- c
		WebsocketInit(r.Context(), svcCtx, c)
		if svcCtx.Config.Mode == service.DevMode || svcCtx.Config.Mode == service.TestMode {
			svcCtx.Config.MsgType = r.Header.Get("X-Content-MsgType")
		}

		go c.WritePump(r.Context())
		go c.ReadPump(r.Context(), svcCtx.Config.MsgType)

	}
}
