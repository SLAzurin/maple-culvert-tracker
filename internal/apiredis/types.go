package apiredis

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type redisInternalKey string

func (k redisInternalKey) ToString() string {
	return string(k)
}
func (k redisInternalKey) Get(rdb *redis.Client) (string, error) {
	return rdb.Get(context.Background(), k.ToString()).Result()
}
func (k redisInternalKey) Set(rdb *redis.Client, v string) error {
	return rdb.Set(context.Background(), k.ToString(), v, 0).Err()
}
func (k redisInternalKey) GetWithDefault(rdb *redis.Client, defaultVal string) string {
	v, err := k.Get(rdb)
	if err == nil {
		return v
	}
	return defaultVal
}

type redisInternalValue string

func (k redisInternalValue) ToString() string {
	return string(k)
}
