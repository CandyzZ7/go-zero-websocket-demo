package test

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/rest/httpx"

	"go-zero-websocket-demo/internal/logic/test"
	"go-zero-websocket-demo/internal/svc"
	"go-zero-websocket-demo/internal/types"

	"go-zero-websocket-demo/pkg"
)

// ping
func PingHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := pkg.Upgrader.Upgrade(w, r, nil)
		if err != nil {
			logc.Errorf(r.Context(), "Error upgrading to WebSocket: %v", err)
			return
		}

		c := &pkg.Connection{Conn: conn}
		h := svcCtx.WSHub
		h.Register <- c

		defer func() {
			h.Unregister <- c
		}()

		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					logc.Errorf(r.Context(), "Error reading message: %v", err)
				}
				break
			}
			var req types.PingReq
			err = json.Unmarshal(message, &req)
			if err != nil {
				httpx.ErrorCtx(r.Context(), w, err)
				return
			}

			l := test.NewPingLogic(r.Context(), svcCtx)
			resp, err := l.Ping(&req)
			if err != nil {
				httpx.ErrorCtx(r.Context(), w, err)
			} else {
				httpx.OkJsonCtx(r.Context(), w, resp)
			}

			bytes, err := json.Marshal(resp)
			if err != nil {
				logc.Error(r.Context(), err)
				return
			}
			h.Broadcast <- bytes
		}
	}
}
