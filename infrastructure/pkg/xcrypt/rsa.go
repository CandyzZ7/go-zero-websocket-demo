package xcrypt

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
)

// RsaGenKey 生成RSA公钥和私钥并保存在对应的目录文件下
// 参数bits: 指定生成的秘钥的长度, 单位: bit
func RsaGenKey(bits int, privatePath, publicPath string) error {
	// 1. 生成私钥文件
	// GenerateKey函数使用随机数据生成器random生成一对具有指定字位数的RSA密钥
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return err
	}
	// 2. MarshalPKCS1PrivateKey将rsa私钥序列化为ASN.1 PKCS#1 DER编码
	derPrivateStream := x509.MarshalPKCS1PrivateKey(privateKey)

	// 3. Block代表PEM编码的结构, 对其进行设置
	block := pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: derPrivateStream,
	}

	// 4. 创建文件
	privateFile, err := os.Create(privatePath)
	defer privateFile.Close()

	if err != nil {
		return err
	}

	// 5. 使用pem编码, 并将数据写入文件中
	err = pem.Encode(privateFile, &block)
	if err != nil {
		return err
	}

	// 1. 生成公钥文件
	publicKey := privateKey.PublicKey
	derPublicStream, err := x509.MarshalPKIXPublicKey(&publicKey)
	if err != nil {
		return err
	}

	block = pem.Block{
		Type:  "RSA PUBLIC KEY",
		Bytes: derPublicStream,
	}

	publicFile, err := os.Create(publicPath)
	defer publicFile.Close()

	if err != nil {
		return err
	}

	// 2. 编码公钥, 写入文件
	err = pem.Encode(publicFile, &block)
	if err != nil {
		panic(err)
		return err
	}
	return nil
}

// EncryptRSA RSA公钥加密
func EncryptRSA(src []byte, filename string) ([]byte, error) {
	// 根据文件名读出文件内容
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	info, _ := file.Stat()
	buf := make([]byte, info.Size())
	file.Read(buf)

	// 从数据中找出pem格式的块
	block, _ := pem.Decode(buf)
	if block == nil {
		return nil, err
	}

	// 解析一个der编码的公钥
	publicKey, err := x509.ParsePKCS1PublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	// 公钥加密
	result, _ := rsa.EncryptPKCS1v15(rand.Reader, publicKey, src)
	return result, nil

}

// DecryptRSA RSA私钥解密
func DecryptRSA(src []byte, filename string) ([]byte, error) {
	// 根据文件名读出内容
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	info, _ := file.Stat()
	buf := make([]byte, info.Size())
	file.Read(buf)

	// 从数据中解析出pem块
	block, _ := pem.Decode(buf)
	if block == nil {
		return nil, err
	}

	// 解析出一个der编码的私钥
	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)

	// 私钥解密
	result, err := rsa.DecryptPKCS1v15(rand.Reader, privateKey, src)
	if err != nil {
		return nil, err
	}
	return result, nil
}
