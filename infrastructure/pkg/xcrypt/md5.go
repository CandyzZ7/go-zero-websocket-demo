package xcrypt

import (
	"crypto/md5"
	"encoding/hex"
	"strings"
)

// EncryptMD5 MD5加密多用于数据完整性校验。
// 例如，下载文件时，网站可能会提供该文件的MD5值，用户下载后通过计算文件的MD5值并与提供的值进行对比，就可以判断文件是否在传输过程中发生了改变。
// 此外，MD5也经常被用于存储用户的密码，例如在数据库中存储用户的MD5密码，即使数据库被泄露，攻击者也无法直接得知用户的真实密码。
func EncryptMD5(encrypt string) string {
	return hex.EncodeToString(byte16ToBytes(md5.Sum([]byte(strings.TrimSpace(encrypt)))))
}

// EqualsEncryptMD5 对比密码是否正确
func EqualsEncryptMD5(target, encryptTarget string) bool {
	md5Target := EncryptMD5(target)
	if md5Target == encryptTarget {
		return true
	}
	return false
}
