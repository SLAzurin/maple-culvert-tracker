package apiredis

import (
	"context"
	"os"

	"github.com/redis/go-redis/v9"
	"github.com/slazurin/maple-culvert-tracker/internal/data"
)

type editableType string

const (
	EditableTypeString         editableType = "string"
	EditableTypeUInt           editableType = "uint"
	EditableTypeDiscordRole    editableType = "discord_role"
	EditableTypeDiscordChannel editableType = "discord_channel"
	EditableTypeSelection      editableType = "selection"
	EditableTypeNone           editableType = "none"
)

type redisInternalKey struct {
	Name         string
	EditableType editableType
	Multiple     bool
}

func (k redisInternalKey) ToString() string {
	return k.Name
}
func (k redisInternalKey) Get(rdb *redis.Client) (string, error) {
	return rdb.Get(context.Background(), os.Getenv(data.EnvVarDiscordGuildID)+"_"+k.ToString()).Result()
}
func (k redisInternalKey) Set(rdb *redis.Client, v string) error {
	return rdb.Set(context.Background(), os.Getenv(data.EnvVarDiscordGuildID)+"_"+k.ToString(), v, 0).Err()
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
