package repository

import (
	"context"

	"go-zero-websocket-demo/internal/model/entity"
)

func GetUserOnlineByAppIDAndUserID(ctx context.Context, appID, userID string) (*entity.UserOnline, error) {
	return &entity.UserOnline{}, nil
}

func UpdateUserOnline(ctx context.Context, userOnline *entity.UserOnline) error {
	return nil
}
