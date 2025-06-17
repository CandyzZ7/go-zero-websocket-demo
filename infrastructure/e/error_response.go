package e

import "github.com/pkg/errors"

type StatusCode struct {
	Code    Code   `json:"code"`
	Message string `json:"message"`
}

// 实现error的接口  然后CodeError继承一下Error方法  CodeError就为error类型的返回值
func (e *StatusCode) Error() string {
	return e.Message
}

func (e *StatusCode) GetCode() Code {
	return e.Code
}

func (e *StatusCode) GetMessage() string {
	return e.Message
}

// ErrorResponse 返回给前端的数据
func (e *StatusCode) ErrorResponse() *StatusCode {
	return &StatusCode{
		Code:    e.Code,
		Message: e.Message,
	}
}

// newStatusCode 提供new方法，任意地方传递参数返回CodeError类型的数据
func newStatusCode(code Code, msg string) *StatusCode {
	return &StatusCode{
		Code:    code,
		Message: msg,
	}
}

// DefaultErrHandler 默认异常状态码函数，只需传递错误信息即可，默认返回code-10001
func DefaultErrHandler(msg string) *StatusCode {
	return &StatusCode{
		Code:    BadRequest.Code,
		Message: msg,
	}
}

// ErrHandler 自定义错误返回函数 错误函数主入口
func ErrHandler(err error) *StatusCode {
	var codeError *StatusCode
	switch {
	// 如果错误类型为CodeError，就返回错误类型的结构体
	case errors.As(err, &codeError):
		return codeError
	default:
		// 系统错误，500 错误提示
		return DefaultErrHandler(err.Error())
	}
}
