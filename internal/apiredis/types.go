package apiredis

import (
	"context"
	"os"

	"github.com/slazurin/maple-culvert-tracker/internal/data"
	redis "github.com/valkey-io/valkey-go"
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
	q := (*rdb).Do(context.Background(), (*rdb).B().Get().Key(os.Getenv(data.EnvVarDiscordGuildID)+"_"+k.ToString()).Build())
	if err := q.Error(); err != nil {
		return "", err
	}
	return q.ToString()
}
func (k redisInternalKey) Set(rdb *redis.Client, v string) error {
	return (*rdb).Do(context.Background(), (*rdb).B().Set().Key(os.Getenv(data.EnvVarDiscordGuildID)+"_"+k.ToString()).Value(v).Build()).Error()
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
