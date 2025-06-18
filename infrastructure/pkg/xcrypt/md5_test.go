package xcrypt

import "testing"

func TestEncryptMD5(t *testing.T) {
	cryptSalt, err := RandomString(RandomNumberLen)
	if err != nil {
		t.Errorf("RandomString error: %v", err)
	}
	md5 := EncryptMD5("Oi3GDHLZOYfP" + cryptSalt)
	t.Log(md5)
}
