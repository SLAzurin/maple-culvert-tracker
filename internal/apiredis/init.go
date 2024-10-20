package apiredis

import (
	"os"

	"github.com/redis/go-redis/v9"
	"github.com/slazurin/maple-culvert-tracker/internal/data"
)

var RedisDB *redis.Client

func init() {
	RedisDB = redis.NewClient(&redis.Options{
		Addr:     os.Getenv(data.EnvVarRedisHost) + ":" + os.Getenv(data.EnvVarRedisPort),
		Password: "", // always use no pw, this is a private redis.
		DB:       0,
	})
}
