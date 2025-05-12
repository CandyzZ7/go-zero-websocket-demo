package {{.pkgName}}

import (
	{{.imports}}
	"go-zero-websocket-demo/pkg/websocketx"
)

type {{.logic}} struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	client *websocketx.Client
}

{{if .hasDoc}}{{.doc}}{{end}}
func New{{.logic}}(ctx context.Context, svcCtx *svc.ServiceContext, client *websocketx.Client) *{{.logic}} {
	return &{{.logic}}{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		client: client,
	}
}

func (l *{{.logic}}) {{.function}}({{.request}}) {{.responseType}} {
	// todo: add your logic here and delete this line

	{{.returnString}}
}
