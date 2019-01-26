package mqtt

import (
	"crypto/rand"
	"github.com/sirupsen/logrus"
	"encoding/base64"
)

// https://elithrar.github.io/article/generating-secure-random-numbers-crypto-rand/

// GenerateRandomBytes returns securely generated random bytes.
// It will fail with a fatal log if the system's secure random
// number generator fails to function correctly
func GenerateRandomBytes(n int) []byte {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		// Note that err == nil only if we read len(b) bytes.
		logrus.Fatal("Could not read random bytes")
	}

	return b
}

// GenerateRandomString returns a URL-safe, base64 encoded
// securely generated random string.
// It will fail with a fatal log if the system's secure random
// number generator fails to function correctly
func GenerateRandomString(s int) string {
	b := GenerateRandomBytes(s)
	return base64.URLEncoding.EncodeToString(b)
}
