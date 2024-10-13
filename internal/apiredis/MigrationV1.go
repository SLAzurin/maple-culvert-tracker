package apiredis

import (
	"context"
	"os"

	"github.com/redis/go-redis/v9"
)

func MigrationV1(rdb *redis.Client) error {
	var migrationV1EnvToKeys = map[string]redisInternalKey{
		"DISCORD_MEMBERS_MAIN_CHANNEL_ID": CONF_DISCORD_MEMBERS_MAIN_CHANNEL_ID,
		"DISCORD_REMINDER_CHANNEL_ID":     CONF_DISCORD_ADMIN_CHANNEL_ID,
		"DISCORD_REMINDER_SUFFIX":         OPTIONAL_CONF_DISCORD_REMINDER_SUFFIX,
		"DISCORD_GUILD_ROLE_ID":           CONF_DISCORD_GUILD_ROLE_IDS,
		"MAPLE_REGION":                    OPTIONAL_CONF_MAPLE_REGION,
		// "CHARTMAKER_HOST":                 "", // This should not be configurable
		// "POSTGRES_USER":                   "", // This should not be configurable
		// "POSTGRES_PASSWORD":               "", // This should not be configurable
		// "POSTGRES_DB":                     "", // This should not be configurable
		// "CLIENT_POSTGRES_PORT":            "", // This should not be configurable
		// "CLIENT_POSTGRES_HOST":            "", // This should not be configurable
		// "BACKEND_HTTP_PORT":               "", // This should continue being a Getenv call
		// "REDIS_HOST":                      "", // This should not be configurable
		// "REDIS_PORT":                      "", // This should not be configurable
		"CULVERT_DUEL_THUMBNAIL_URL": OPTIONAL_CONF_CULVERT_DUEL_THUMBNAIL_URL,
	}
	for k, v := range migrationV1EnvToKeys {
		err := rdb.Set(context.Background(), v.ToString(), os.Getenv(k), 0).Err()
		if err != nil {
			return err
		}
	}
	return nil
}