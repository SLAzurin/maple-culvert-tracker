package apiredis

import (
	"log"
	"strconv"

	"github.com/redis/go-redis/v9"
)

const CurrentVersion = 1

var migrationTable = map[int]func(rdb *redis.Client) error{
	1: MigrationV1,
}

func Migrate(rdb *redis.Client) error {
	v, err := DATA_REDIS_VERSION.Get(rdb)
	if err != nil && err != redis.Nil {
		log.Println("Failed to get redis data version "+DATA_REDIS_VERSION.Name, err)
		return err
	}
	if v == "" {
		v = "0"
	}

	i, err := strconv.Atoi(v)
	if err != nil {
		log.Println("Failed to convert redis data version to int "+v, err)
		log.Println("Treating as Version 0")
		i = 0
	}
	for i < CurrentVersion {
		log.Println("Running Migration from version " + strconv.Itoa(i) + " to " + strconv.Itoa(i+1))
		err := migrationTable[i+1](rdb)
		if err != nil {
			log.Println("Failed to run Migration from version "+strconv.Itoa(i)+" to "+strconv.Itoa(i+1), err)
			return err
		}
		err = DATA_REDIS_VERSION.Set(rdb, strconv.Itoa(i+1))
		if err != nil {
			log.Println("Failed to set redis data version "+DATA_REDIS_VERSION.Name, err)
			return err
		}
		i++
	}
	return nil
}
