package randomidgen

import (
	"math/rand"
	"strings"
	"time"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// init seeds the random number generator with the current time.
func init() {
	rand.NewSource(time.Now().UnixNano())
}

// RandomID generates a random string of the specified length.
func RandomID(length int) string {
	if length <= 0 {
		return ""
	}
	var sb strings.Builder
	sb.Grow(length)
	for i := 0; i < length; i++ {
		randomChar := charset[rand.Intn(len(charset))]
		sb.WriteByte(randomChar)
	}
	return sb.String()
}
