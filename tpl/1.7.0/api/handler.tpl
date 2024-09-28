package {{.PkgName}}

import (
	"net/http"
	"encoding/json"

    "github.com/gorilla/websocket"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/rest/httpx"
	{{.ImportPackages}}
	"go-zero-websocket/pkg"
)

{{if .HasDoc}}{{.Doc}}{{end}}
func {{.HandlerName}}(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
        conn, err := pkg.Upgrader.Upgrade(w, r, nil)
                if err != nil {
                    logc.Errorf(r.Context(),"Error upgrading to WebSocket: %v", err)
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
                    logc.Errorf(r.Context(),"Error reading message: %v", err)
                }
                break
            }
            {{if .HasRequest}}var req types.{{.RequestType}}
            err = json.Unmarshal(message, &req)
            if err != nil {
                httpx.ErrorCtx(r.Context(), w, err)
                return
            }

            {{end}}l := {{.LogicName}}.New{{.LogicType}}(r.Context(), svcCtx)
            {{if .HasResp}}resp, {{end}}err := l.{{.Call}}({{if .HasRequest}}&req{{end}})
            if err != nil {
                httpx.ErrorCtx(r.Context(), w, err)
            } else {
                {{if .HasResp}}httpx.OkJsonCtx(r.Context(), w, resp){{else}}httpx.Ok(w){{end}}
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
