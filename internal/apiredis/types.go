package apiredis

import (
	"context"
	"os"

	"github.com/redis/go-redis/v9"
)

type editableType string

const (
	editableTypeString         editableType = "string"
	editableTypeUInt           editableType = "uint"
	editableTypeDiscordRole    editableType = "discord_role"
	editableTypeDiscordChannel editableType = "discord_channel"
	editableTypeSelection      editableType = "selection"
	editableTypeNone           editableType = "none"
)

type redisInternalKey struct {
	Name         string
	EditableType editableType
}

func (k redisInternalKey) ToString() string {
	return k.Name
}
func (k redisInternalKey) Get(rdb *redis.Client) (string, error) {
	return rdb.Get(context.Background(), os.Getenv("DISCORD_GUILD_ID")+"_"+k.ToString()).Result()
}
func (k redisInternalKey) Set(rdb *redis.Client, v string) error {
	return rdb.Set(context.Background(), os.Getenv("DISCORD_GUILD_ID")+"_"+k.ToString(), v, 0).Err()
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
