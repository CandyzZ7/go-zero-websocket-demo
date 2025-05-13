package {{.PkgName}}

import (
	"net/http"
	"encoding/json"

    "github.com/gorilla/websocket"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/rest/httpx"
	{{.ImportPackages}}
	"go-zero-websocket-demo/pkg/websocketx"
)

{{if .HasDoc}}{{.Doc}}{{end}}
func {{.HandlerName}}(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
        conn, err := websocketx.Upgrader.Upgrade(w, r, nil)
		if err != nil {
			logc.Errorf(r.Context(), "error upgrading to WebSocket: %v", err)
			return
		}

		currentTime := uint64(time.Now().Unix())
		h := svcCtx.WSHub
		c := websocketx.NewClient(h, conn.RemoteAddr().String(), conn, currentTime)
		h.Register <- c

		go c.WritePump()

		go func(client *websocketx.Client) {
		    defer func() {
        		if r := recover(); r != nil {
        			logx.Info("write stop", string(debug.Stack()), r)
        		}
        	}()
        	defer func() {
        		logx.Info("client %s disconnected", client.Addr)
        		close(c.Send)
        	}()
			for {
				_, message, err := conn.ReadMessage()
				if err != nil {
					if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
						logx.Infof("websocket was closed unexpectedly: %v", err)
					}
					break
				}
				logc.Infof(r.Context(), "received message: %s", message)
                {{if .HasRequest}}var req types.{{.RequestType}}
                err = json.Unmarshal(message, &req)
                if err != nil {
                    httpx.ErrorCtx(r.Context(), w, err)
                    return
                }

                {{end}}l := {{.LogicName}}.New{{.LogicType}}(r.Context(), svcCtx, client)
                {{if .HasResp}}resp, {{end}}err := l.{{.Call}}({{if .HasRequest}}&req{{end}})
			    if err != nil {
				    logc.Error(r.Context(), err)
				    continue
			    }

                bytes, err := json.Marshal(resp)
                if err != nil {
                    logc.Error(r.Context(), err)
                    return
                }
            	c.Send <- bytes
			}
		}(c)
	}
}
