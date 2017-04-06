package mass

import (
	"github.com/garyburd/redigo/redis"
)

var unlockScript = redis.NewScript(1, `
	if redis.call("get", KEYS[1]) == ARGV[1]
	then
		return redis.call("del", KEYS[1])
	else
		return 0
	end
`)

func Lock(rc redis.Conn, key string, secret string, ttl uint64) (bool, error) {
	status, err := redis.String(rc.Do("SET", key, secret, "EX", int64(ttl), "NX"))
	if err == redis.ErrNil {
		// The lock was not successful, it already exists.
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return status == "OK", nil
}

func Unlock(rc redis.Conn, key string, secret string) error {
	_, err := unlockScript.Do(rc, key, secret)
	return err
}
