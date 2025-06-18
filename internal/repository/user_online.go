package repository

import (
	"context"
	"encoding/json"

	"github.com/zeromicro/go-zero/core/stores/redis"

	"go-zero-websocket-demo/common/rediskey"
	"go-zero-websocket-demo/infrastructure/pkg/redisx"
	"go-zero-websocket-demo/internal/model"
)

const userOnlineCacheTime = 24 * 60 * 60

func GetUserOnlineByAppIDAndUserID(ctx context.Context, appID, userID string) (*model.UserOnline, error) {
	redisClient := redisx.GetRedisClient()
	key := rediskey.UserOnline.WithParams(appID, userID)
	data, err := redisClient.Get(key)
	if err != nil {
		return nil, err
	}
	if data == "" {
		return nil, redis.Nil
	}
	userOnline := &model.UserOnline{}

	err = json.Unmarshal([]byte(data), userOnline)
	if err != nil {
		return nil, err
	}
	return userOnline, nil
}

func UpdateUserOnline(ctx context.Context, userOnline *model.UserOnline) error {
	redisClient := redisx.GetRedisClient()
	key := rediskey.UserOnline.WithParams(userOnline.AppID, userOnline.UserID)
	data, err := json.Marshal(userOnline)
	if err != nil {
		return err
	}
	err = redisClient.SetexCtx(ctx, key, string(data), userOnlineCacheTime)
	if err != nil {
		return err
	}
	return nil
}
