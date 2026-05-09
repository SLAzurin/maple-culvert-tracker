package apiredis

import (
	redis "github.com/valkey-io/valkey-go"
)

func MigrationV5(vk *redis.Client) error {
	return OPTIONAL_CONF_MONTHLY_IMPROVEMENT_THRESHOLD.Set(vk, "10")
}
