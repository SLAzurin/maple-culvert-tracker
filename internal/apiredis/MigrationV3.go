package apiredis

import (
	redis "github.com/valkey-io/valkey-go"
)

func MigrationV3(rdb *redis.Client) error {
	return OPTIONAL_CONF_SUBMIT_SCORES_SHOW_RATS.Set(rdb, "false")
}
