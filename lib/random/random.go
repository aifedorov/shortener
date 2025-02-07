package random

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
)

func GenRandomString(s string, size int) (string, error) {
	if size <= 0 || size > sha256.Size {
		return "", fmt.Errorf("random: invalid size %d, must be > 0 and <= %d", size, sha256.Size)
	}
	hash := sha256.Sum256([]byte(s))
	return base64.RawURLEncoding.EncodeToString(hash[:size]), nil
}
