package apiredis

import (
	redis "github.com/valkey-io/valkey-go"
)

func MigrationV4(vk *redis.Client) error {
	return OPTIONAL_CONF_SANDBAG_THRESHOLD.Set(vk, "0.7")
}
