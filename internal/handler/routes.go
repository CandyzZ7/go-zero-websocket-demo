// Code generated by goctl. DO NOT EDIT.
package handler

import (
	"net/http"

	test "go-zero-websocket-demo/internal/handler/test"
	"go-zero-websocket-demo/internal/svc"

	"github.com/zeromicro/go-zero/rest"
)

func RegisterHandlers(server *rest.Server, serverCtx *svc.ServiceContext) {
	server.AddRoutes(
		[]rest.Route{
			{
				// ping
				Method:  http.MethodGet,
				Path:    "/ping",
				Handler: test.PingHandler(serverCtx),
			},
		},
	)
}
