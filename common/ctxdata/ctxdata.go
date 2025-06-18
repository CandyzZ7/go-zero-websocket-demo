package ctxdata

import (
	"context"
)

var CtxKeyJwtUserId = "jwtUserId"
var CtxKeyJwtAPPId = "jwtAPPId"
var CtxKeyJwtChannelId = "jwtChannel"
var CtxKeyJwtGameId = "jwtGameId"
var CtxKeyJwtGamePackageId = "jwtGamePackageId"

func GetUidFromCtx(ctx context.Context) string {
	if uid, ok := ctx.Value(CtxKeyJwtUserId).(string); ok {
		return uid
	}
	return ""
}

func GetAppIdFromCtx(ctx context.Context) string {
	if appId, ok := ctx.Value(CtxKeyJwtAPPId).(string); ok {
		return appId
	}
	return ""
}

func GetChannelIdFromCtx(ctx context.Context) string {
	if channel, ok := ctx.Value(CtxKeyJwtChannelId).(string); ok {
		return channel
	}
	return ""
}

func GetGameIdFromCtx(ctx context.Context) string {
	if gameId, ok := ctx.Value(CtxKeyJwtGameId).(string); ok {
		return gameId
	}
	return ""
}

func GetGamePackageIdFromCtx(ctx context.Context) string {
	if gamePackageId, ok := ctx.Value(CtxKeyJwtGamePackageId).(string); ok {
		return gamePackageId
	}
	return ""
}
