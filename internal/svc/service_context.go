package svc

import (
	"github.com/zeromicro/go-zero/core/stores/redis"

	"go-zero-websocket-demo/infrastructure/pkg/redisx"
	"go-zero-websocket-demo/internal/config"
)

type ServiceContext struct {
	Config config.Config
	RDB    *redis.Redis
}

func NewServiceContext(c config.Config) *ServiceContext {
	rdb := redisx.MustNewRedisClient(c.RedisConf)

	return &ServiceContext{
		Config: c,
		RDB:    rdb,
	}
}
