package xcrypt

import (
	"bytes"
	"encoding/base64"

	"github.com/zeromicro/go-zero/core/codec"
)

const (
	AESKey = "sakuragamesecret"
	AESIv  = "secretsakuragame"
)

func EncryptAESByCBC(key, iv, encrypt string) (string, error) {
	encrypted, err := CbcEncryptBase64(key, iv, base64.StdEncoding.EncodeToString([]byte(encrypt)))
	if err != nil {
		return "", err
	}
	return encrypted, nil
}

func DecryptAESByCBC(key, iv, encrypt string) (string, error) {
	decrypted, err := CbcDecryptBase64(key, iv, encrypt)
	if err != nil {
		return "", err
	}
	decodeString, err := base64.StdEncoding.DecodeString(decrypted)
	if err != nil {
		return "", err
	}
	return string(decodeString), nil
}

func EncryptAESByECB(key, encrypt string) (string, error) {
	data, err := codec.EcbEncrypt([]byte(key), []byte(encrypt))
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(data), nil
}

func DecryptAESByECB(key, encrypt string) (string, error) {
	originalData, err := base64.StdEncoding.DecodeString(encrypt)
	if err != nil {
		return "", err
	}
	data, err := codec.EcbDecrypt([]byte(key), originalData)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func EncryptAESBase64ByECB(key, encrypt string) (string, error) {
	data, err := codec.EcbEncryptBase64(key, encrypt)
	if err != nil {
		return "", err
	}
	return data, nil
}

func DecryptAESBase64ByECB(key, encrypt string) (string, error) {
	data, err := codec.EcbDecryptBase64(key, encrypt)
	if err != nil {
		return "", err
	}
	return data, nil
}

func pkcs5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func pkcs5Unpadding(src []byte, blockSize int) ([]byte, error) {
	length := len(src)
	unpadding := int(src[length-1])
	if unpadding >= length || unpadding > blockSize {
		return nil, codec.ErrPaddingSize
	}

	return src[:length-unpadding], nil
}
