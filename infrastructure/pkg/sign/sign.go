package commonmethod

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/zeromicro/go-zero/core/codec"
	"github.com/zeromicro/go-zero/core/iox"
)

const (
	SHA256 = "hmac_sha256"
)

func SignAndConnectWebSocket(mode, method, wsURL, key, pubKey string, reqBody []byte) (*websocket.Conn, error) {
	reader := bytes.NewBuffer(reqBody)
	r, err := http.NewRequest(method, wsURL, reader)
	if err != nil {
		return nil, err
	}
	// 计算签名所需的内容
	timestamp := time.Now().Unix()
	contentOfSign := strings.Join([]string{
		strconv.FormatInt(timestamp, 10),
		"GET", // WebSocket握手使用GET方法
		r.URL.Path,
		r.URL.RawQuery,
		computeBodySignature(r),
	}, "\n")

	sign := hs256([]byte(key), contentOfSign)
	content := strings.Join([]string{
		"type=" + mode,
		fmt.Sprintf("key=%s", base64.StdEncoding.EncodeToString([]byte(key))),
		"time=" + strconv.FormatInt(timestamp, 10),
	}, "; ")

	encrypter, err := codec.NewRsaEncrypter([]byte(pubKey))
	if err != nil {
		return nil, err
	}

	output, err := encrypter.Encrypt([]byte(content))
	if err != nil {
		return nil, err
	}

	encryptedContent := base64.StdEncoding.EncodeToString(output)

	// 创建WebSocket连接
	headers := make(map[string][]string)
	headers["X-Content-Security"] = []string{strings.Join([]string{
		fmt.Sprintf("key=%s", key),
		"secret=" + encryptedContent,
		"signature=" + sign,
	}, "; ")}
	headers["X-Content-MsgType"] = []string{"json"}

	dialer := &websocket.Dialer{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // 仅用于测试环境
	}

	conn, _, err := dialer.Dial(wsURL, headers)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func hs256(key []byte, body string) string {
	h := hmac.New(sha256.New, key)
	io.WriteString(h, body)
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func computeBodySignature(r *http.Request) string {
	var dup io.ReadCloser
	r.Body, dup = iox.DupReadCloser(r.Body)
	sha := sha256.New()
	io.Copy(sha, r.Body)
	r.Body = dup
	return fmt.Sprintf("%x", sha.Sum(nil))
}

func HMAC(sessionKey, rawData []byte, method string) string {
	// 创建一个新的HMAC哈希对象，使用SHA256算法
	var h hash.Hash
	switch method {
	case SHA256:
		h = hmac.New(sha256.New, sessionKey)
	default:
		h = hmac.New(sha1.New, sessionKey)
	}
	// 将原始数据写入哈希对象
	h.Write(rawData)
	// 计算哈希值
	signature := h.Sum(nil)
	// 将哈希值转换为十六进制字符串
	return hex.EncodeToString(signature)
}

func SendWebSocketMessage(conn *websocket.Conn, messageType int, data []byte) error {
	return conn.WriteMessage(messageType, data)
}

func ReceiveWebSocketMessage(conn *websocket.Conn) (int, []byte, error) {
	return conn.ReadMessage()
}
