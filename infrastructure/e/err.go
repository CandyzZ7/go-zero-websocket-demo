package e

var (
	OK          = newStatusCode(OKCode, OKCode.String())
	BadRequest  = newStatusCode(BadRequestCode, BadRequestCode.String())
	SystemError = newStatusCode(SystemErrorCode, SystemErrorCode.String())
	ParseError  = newStatusCode(ParseErrorCode, ParseErrorCode.String())
)

var (
	ErrTokenVerifyFail = newStatusCode(ErrTokenVerifyFailCode, ErrTokenVerifyFailCode.String())
)

var (
	NotLoggedIn = newStatusCode(NotLoggedInCode, NotLoggedInCode.String())
)

var (
	NotFoundRoute = newStatusCode(NotFoundRouteCode, NotFoundRouteCode.String())
)
