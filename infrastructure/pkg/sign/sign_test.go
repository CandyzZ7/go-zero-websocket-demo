package commonmethod

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
)

const (
	pubKey = `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAtuhWQ4pxOdK976PumB3b
0I2dz00wH9ss2ebdb04RbtYui0/BVoR7DNsXs6TGGczInHF/IoNTfcLhAoCy2MAV
eaUT3iX0AzyE+Aj9DBdkjMBmy9GpHOWHMNuzmHyDnyijVeWA8KUqqmm4yKviKe6P
cnRpcwY5+DbDXqtwrIOSt+qOB7xb8+sUXU/gXW+Voeop8h0LcfOPTFdt0Eo1WwEz
ZXXiNc3A5/N+V42+wxqjmRv3h83OePKqKmCUF39n72Ch9WXUkMWkiKwadIX0Jd9W
kzB7Dv9hBKiRu6E/OCZQ7f8ySC8kbCwDMrxlsSrsfU7af+j8pe+hR6YpxWu4YEWO
8QIDAQAB
-----END PUBLIC KEY-----
`
)

func TestSignAndConnectWebSocket(t *testing.T) {
	conn, err := SignAndConnectWebSocket(
		"0",
		http.MethodGet,
		"ws://127.0.0.1:8888/ws",
		"ydgame",
		pubKey,
		[]byte(""),
	)

	assert.Nil(t, err)
	assert.NotNil(t, conn)

	if conn != nil {
		defer conn.Close()

		// 发送测试消息
		err = SendWebSocketMessage(conn, websocket.TextMessage, []byte(`{
    "seq": "",
    "cmd": "test.ping",
    "data": {
        "msg": "yd"
    }
}`))
		assert.Nil(t, err)

		// 接收响应
		messageType, p, err := ReceiveWebSocketMessage(conn)
		assert.Nil(t, err)
		assert.NotEmpty(t, p)

		fmt.Printf("Received message: %s (type: %d)\n", string(p), messageType)
	}
}
