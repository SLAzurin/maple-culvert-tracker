package main

import (
	"fmt"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	_ "github.com/joho/godotenv/autoload"
	"github.com/slazurin/maple-culvert-tracker/internal/data"
)

func main() {
	mode := 0
	if os.Getenv(gin.EnvGinMode) != gin.ReleaseMode {
		mode = 1
	}
	displayName := "Backdoor user hehe ecksdee"
	claims := &data.MCTClaims{
		DiscordUsername: displayName,
		DiscordServerID: os.Getenv(data.EnvVarDiscordGuildID),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(4 * time.Hour)),
		},
		DevMode: mode,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(os.Getenv(data.EnvVarJWTSecret)))

	fmt.Println(tokenString)
}
