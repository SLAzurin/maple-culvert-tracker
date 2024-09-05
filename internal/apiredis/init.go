package apiredis

import (
	"os"

	"github.com/redis/go-redis/v9"
)

var RedisDB *redis.Client

func init() {
	RedisDB = redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_HOST") + ":" + os.Getenv("REDIS_PORT"),
		Password: "", // always use no pw, this is a private redis.
		DB:       0,
	})
}
