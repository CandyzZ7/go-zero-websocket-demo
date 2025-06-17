package e

var (
	OK          = newStatusCode(OKCode, OKCode.String())
	BadRequest  = newStatusCode(BadRequestCode, BadRequestCode.String())
	SystemError = newStatusCode(SystemErrorCode, SystemErrorCode.String())
	ParseError  = newStatusCode(ParseErrorCode, ParseErrorCode.String())
)

var (
	NotFoundRoute = newStatusCode(NotFoundRouteCode, NotFoundRouteCode.String())
)
