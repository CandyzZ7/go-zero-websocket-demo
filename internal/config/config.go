package config

import (
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/rest"

	"go-zero-websocket-demo/infrastructure/pkg/ormengine"
)

type Config struct {
	rest.RestConf
	RedisConf redis.RedisConf
	MsgType   string
	MysqlConf ormengine.MysqlConf
	TokenConf TokenConf
}

type TokenConf struct {
	AccessSecret  string
	AccessExpire  int64
	RefreshSecret string
	RefreshExpire int64
	RefreshAfter  int64
}
