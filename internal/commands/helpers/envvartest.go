package helpers

import (
	"log"
	"os"
)

func EnvVarsTest() {
	log.Println("Ensuring all env vars are set")

	if os.Getenv("DISCORD_TOKEN") == "" {
		log.Fatalln("DISCORD_TOKEN is not set")
	}
	if os.Getenv("DISCORD_GUILD_ID") == "" {
		log.Fatalln("DISCORD_GUILD_ID is not set")
	}
	if os.Getenv("JWT_SECRET") == "" {
		log.Fatalln("JWT_SECRET is not set")
	}
	if os.Getenv("CHARTMAKER_HOST") == "" {
		log.Fatalln("CHARTMAKER_HOST is not set")
	}
	if os.Getenv("POSTGRES_USER") == "" {
		log.Fatalln("POSTGRES_USER is not set")
	}
	if os.Getenv("POSTGRES_PASSWORD") == "" {
		log.Fatalln("POSTGRES_PASSWORD is not set")
	}
	if os.Getenv("CLIENT_POSTGRES_PORT") == "" {
		log.Fatalln("CLIENT_POSTGRES_PORT is not set")
	}
	if os.Getenv("CLIENT_POSTGRES_HOST") == "" {
		log.Fatalln("CLIENT_POSTGRES_HOST is not set")
	}
	// In theory, this should not be needed. But should be set
	// if os.Getenv("FRONTEND_URL") == "" {
	// 	log.Fatalln("FRONTEND_URL is not set")
	// }
	log.Println("Frontend URL missing but should be set. This does not break anything but it will work anyway without it.")
	if os.Getenv("REDIS_HOST") == "" {
		log.Fatalln("REDIS_HOST is not set")
	}
	if os.Getenv("REDIS_PORT") == "" {
		log.Fatalln("REDIS_PORT is not set")
	}

	log.Println("All env vars are set, continuing...")
}
