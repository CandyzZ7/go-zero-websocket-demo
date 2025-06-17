package rediskey

type RedisKey string

const WebSocketKey RedisKey = "websocket"

func (key RedisKey) WithSymbol(symbol string) string {
	return string(key) + "_" + symbol
}

func (key RedisKey) WithParams(params ...string) string {
	if len(params) == 0 {
		return string(key)
	}
	k := string(key)
	for _, v := range params {
		k += ":" + v
	}
	return k
}
