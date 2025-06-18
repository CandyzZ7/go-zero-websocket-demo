package xcrypt

import (
	"fmt"
	"testing"
)

func TestEncryptASEBase64ByECB(t *testing.T) {
	encryptedData, err := EncryptAESByECB(AESKey, "HELLO")
	if err != nil {
		t.Error(err)
	}
	data, err := DecryptAESByECB(AESKey, encryptedData)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(data)
}

// TestEncryptASEBase64ByCBC 测试 CBC 加密和解密
func TestEncryptASEBase64ByCBC(t *testing.T) {
	encryptedData, err := EncryptAESByCBC(AESKey, AESIv, "HELLO")
	if err != nil {
		t.Error(err)
	}
	data, err := DecryptAESByCBC(AESKey, AESIv, encryptedData)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(data)
}
