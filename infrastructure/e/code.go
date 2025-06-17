package e

//go:generate stringer -type Code -linecomment

type Code int

const (
	// OKCode 成功
	OKCode Code = 0 // OK
	// BadRequestCode 错误请求
	BadRequestCode  Code = -1 // Server Error
	SystemErrorCode Code = -2 // System Error
	ParseErrorCode  Code = -3 // Parse Error
)

const (
	// NotFoundRouteCode  未找到路由
	NotFoundRouteCode Code = 4001 // Not Found Route
)
