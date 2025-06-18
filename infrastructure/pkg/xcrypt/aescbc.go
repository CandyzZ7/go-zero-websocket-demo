package xcrypt

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"

	"github.com/pkg/errors"
	"github.com/zeromicro/go-zero/core/logx"
)

// ErrPaddingSize indicates bad padding size.
var ErrPaddingSize = errors.New("padding size error")

// Cbc struct to hold block and block size
type cbc struct {
	b         cipher.Block
	blockSize int
	iv        []byte
}

// NewCBC returns a new CBC instance
func newCBC(b cipher.Block, iv []byte) *cbc {
	return &cbc{
		b:         b,
		blockSize: b.BlockSize(),
		iv:        iv,
	}
}

// cbcEncrypter structure
type cbcEncrypter cbc

// NewCBCEncrypter returns a CBC encrypter.
func NewCBCEncrypter(b cipher.Block, iv []byte) cipher.BlockMode {
	return (*cbcEncrypter)(newCBC(b, iv))
}

// BlockSize returns the mode's block size.
func (x *cbcEncrypter) BlockSize() int { return x.blockSize }

// CryptBlocks encrypts a number of blocks.
func (x *cbcEncrypter) CryptBlocks(dst, src []byte) {
	if len(src)%x.blockSize != 0 {
		logx.Error("crypto/cipher: input not full blocks")
		return
	}
	if len(dst) < len(src) {
		logx.Error("crypto/cipher: output smaller than input")
		return
	}

	// Create an initialization vector from the iv field
	iv := make([]byte, x.blockSize)
	copy(iv, x.iv)

	for len(src) > 0 {
		// XOR the input block with the IV
		for i := 0; i < x.blockSize; i++ {
			src[i] ^= iv[i]
		}

		x.b.Encrypt(dst, src[:x.blockSize])

		// Update IV to the current ciphertext block for next block
		iv = dst[:x.blockSize]

		src = src[x.blockSize:]
		dst = dst[x.blockSize:]
	}
}

// cbcDecrypter structure
type cbcDecrypter cbc

// NewCBCDecrypter returns a CBC decrypter.
func NewCBCDecrypter(b cipher.Block, iv []byte) cipher.BlockMode {
	return (*cbcDecrypter)(newCBC(b, iv))
}

// BlockSize returns the mode's block size.
func (x *cbcDecrypter) BlockSize() int {
	return x.blockSize
}

// CryptBlocks decrypts a number of blocks.
func (x *cbcDecrypter) CryptBlocks(dst, src []byte) {
	if len(src)%x.blockSize != 0 {
		logx.Error("crypto/cipher: input not full blocks")
		return
	}
	if len(dst) < len(src) {
		logx.Error("crypto/cipher: output smaller than input")
		return
	}

	iv := make([]byte, x.blockSize)
	copy(iv, x.iv)

	for len(src) > 0 {
		// 保存当前密文块以用于下一个块的 XOR
		cipherBlock := src[:x.blockSize]

		// 解密当前块
		x.b.Decrypt(dst, cipherBlock)

		// 将解密后的数据与上一个密文块（IV）进行 XOR 操作，恢复原始明文
		for i := 0; i < x.blockSize; i++ {
			dst[i] ^= iv[i]
		}

		// 更新 IV 为当前密文块
		copy(iv, cipherBlock)

		src = src[x.blockSize:]
		dst = dst[x.blockSize:]
	}
}

// CbcDecrypt decrypts src with the given key and IV.
func CbcDecrypt(key, iv, src []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		logx.Errorf("Decrypt key error: % x", key)
		return nil, err
	}

	decrypter := NewCBCDecrypter(block, iv)
	decrypted := make([]byte, len(src))
	decrypter.CryptBlocks(decrypted, src)

	return pkcs5Unpadding(decrypted, decrypter.BlockSize())
}

// CbcDecryptBase64 decrypts base64 encoded src with the given base64 encoded key and IV.
// The returned string is also base64 encoded.
func CbcDecryptBase64(key, iv, src string) (string, error) {
	keyBytes, err := getKeyBytes(key)
	if err != nil {
		return "", err
	}

	ivBytes, err := base64.StdEncoding.DecodeString(iv)
	if err != nil {
		return "", err
	}

	encryptedBytes, err := base64.StdEncoding.DecodeString(src)
	if err != nil {
		return "", err
	}

	decryptedBytes, err := CbcDecrypt(keyBytes, ivBytes, encryptedBytes)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(decryptedBytes), nil
}

// CbcEncrypt encrypts src with the given key and IV.
func CbcEncrypt(key, iv, src []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		logx.Errorf("Encrypt key error: % x", key)
		return nil, err
	}

	padded := pkcs5Padding(src, block.BlockSize())
	crypted := make([]byte, len(padded))
	encrypter := NewCBCEncrypter(block, iv)
	encrypter.CryptBlocks(crypted, padded)

	return crypted, nil
}

// CbcEncryptBase64 encrypts base64 encoded src with the given base64 encoded key and IV.
// The returned string is also base64 encoded.
func CbcEncryptBase64(key, iv, src string) (string, error) {
	keyBytes, err := getKeyBytes(key)
	if err != nil {
		return "", err
	}

	ivBytes, err := base64.StdEncoding.DecodeString(iv)
	if err != nil {
		return "", err
	}

	srcBytes, err := base64.StdEncoding.DecodeString(src)
	if err != nil {
		return "", err
	}

	encryptedBytes, err := CbcEncrypt(keyBytes, ivBytes, srcBytes)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(encryptedBytes), nil
}

func getKeyBytes(key string) ([]byte, error) {
	if len(key) <= 32 {
		return []byte(key), nil
	}

	keyBytes, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		return nil, err
	}

	return keyBytes, nil
}
