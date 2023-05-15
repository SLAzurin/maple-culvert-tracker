package main

import (
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	_ "github.com/joho/godotenv/autoload"
	"github.com/slazurin/maple-culvert-tracker/internal/data"
)

func main() {
	displayName := "Backdoor user hehe ecksdee"
	claims := &data.MCTClaims{
		DiscordUsername: displayName,
		DiscordServerID: os.Getenv("DISCORD_GUILD_ID"),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(4 * time.Hour)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(os.Getenv("JWT_SECRET")))

	fmt.Println(tokenString)
}
