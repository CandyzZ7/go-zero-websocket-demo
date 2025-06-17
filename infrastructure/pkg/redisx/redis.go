package redisx

import (
	"sync"

	"github.com/zeromicro/go-zero/core/stores/redis"
)

var (
	redisClient *redis.Redis
	once        sync.Once
)

func NewRedisClient(conf redis.RedisConf, opts ...redis.Option) (*redis.Redis, error) {
	var err error
	once.Do(func() {
		redisClient, err = redis.NewRedis(conf, opts...)
		if err != nil {
			return
		}
	})
	return redisClient, err
}
func MustNewRedisClient(conf redis.RedisConf, opts ...redis.Option) *redis.Redis {
	once.Do(func() {
		redisClient = redis.MustNewRedis(conf, opts...)
	})
	return redisClient
}

func GetRedisClient() *redis.Redis {
	return redisClient
}
