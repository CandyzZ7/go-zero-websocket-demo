package handler

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest"

	"go-zero-websocket-demo/internal/svc"
)

func RegisterHandlers(server *rest.Server, serverCtx *svc.ServiceContext) {
	server.AddRoutes(
		[]rest.Route{
			{
				// ping
				Method:  http.MethodGet,
				Path:    "/ws",
				Handler: WsHandler(serverCtx),
			},
		},
		rest.WithSignature(serverCtx.Config.Signature),
	)
}
