package svc

import (
	"github.com/zeromicro/go-zero/core/stores/redis"

	"go-zero-websocket-demo/internal/config"
	"go-zero-websocket-demo/pkg/websocketx"
)

type ServiceContext struct {
	Config config.Config
	WSHub  *websocketx.Hub
	RDB    *redis.Redis
}

func NewServiceContext(c config.Config) *ServiceContext {
	rdb := redis.MustNewRedis(c.RedisConf)
	wsHub := websocketx.NewHub()
	go wsHub.Run()

	return &ServiceContext{
		Config: c,
		WSHub:  wsHub,
		RDB:    rdb,
	}
}
