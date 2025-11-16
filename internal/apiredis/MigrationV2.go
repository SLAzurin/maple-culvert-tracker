package apiredis

import (
	redis "github.com/valkey-io/valkey-go"
)

func MigrationV2(rdb *redis.Client) error {
	return OPTIONAL_CONF_SUBMIT_SCORES_SHOW_SANDBAGGERS.Set(rdb, "false")
}
