package mac

import (
	"crypto/hmac"
	"crypto/sha256"
)

func Tag(key, body []byte) [32]byte {
	mac := hmac.New(sha256.New, key)
	mac.Write(body)
	var out [32]byte
	copy(out[:], mac.Sum(nil))
	return out
}

func Verify(key, body, tag []byte) bool {
	if len(key) == 0 || len(tag) != sha256.Size {
		return false
	}
	got := Tag(key, body)
	return hmac.Equal(got[:], tag)
}
