package redis

import "time"

func ttlFromEnd(ends time.Time, buf time.Duration) time.Duration {
	now := time.Now().UTC()
	d := ends.Sub(now) + buf
	if d < buf {
		d = buf
	}
	return d
}
