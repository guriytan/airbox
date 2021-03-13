package encryption

import "time"

// exp return expiration time in second
func exp(ttl time.Duration) int64 {
	return time.Now().Add(ttl).Unix()
}
