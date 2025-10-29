package main

import (
	"context"
	"flag"
	"log"
	"os"

	_ "github.com/joho/godotenv/autoload"
	"github.com/slazurin/maple-culvert-tracker/internal/apiredis"
	"github.com/slazurin/maple-culvert-tracker/internal/commands/helpers"
	"github.com/slazurin/maple-culvert-tracker/internal/data"
)

func main() {
	fromID := flag.String("from", "", "from discord id")
	toID := flag.String("to", "", "to discord id")

	flag.Parse()

	if *fromID == "" {
		panic("from flag is required")
	}
	if *toID == "" {
		*toID = os.Getenv(data.EnvVarDiscordGuildID)
	}
	if *toID == "" {
		panic("to flag is required or set env var DISCORD_GUILD_ID")
	}
	log.Println("Converting discord id from", *fromID, "to", *toID)

	helpers.EnvVarsTest()
	defer (*apiredis.RedisDB).Close()

	keysRaw, err := (*apiredis.RedisDB).Do(context.Background(), (*apiredis.RedisDB).B().Keys().Pattern(*fromID+"_*").Build()).ToArray()
	if err != nil {
		panic(err)
	}
	var keys []string = make([]string, len(keysRaw))
	for i, key := range keysRaw {
		keyVal, err := key.ToString()
		if err != nil {
			panic(err)
		}
		keys[i] = keyVal
	}

	log.Println("Found", len(keys), "keys to convert", keys)

	for _, v := range keys {
		var oldKey string = v
		newKey := *toID + oldKey[len(*fromID):]
		log.Println("Renaming", oldKey, "to", newKey)
		err := (*apiredis.RedisDB).Do(context.Background(), (*apiredis.RedisDB).B().Rename().Key(oldKey).Newkey(newKey).Build()).Error()
		if err != nil {
			log.Println("Failed to rename" + oldKey + " to " + newKey)
			panic(err)
		}
	}
}
