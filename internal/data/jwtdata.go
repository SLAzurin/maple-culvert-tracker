package data

import (
	"github.com/golang-jwt/jwt/v5"
)

type MCTClaims struct {
	jwt.RegisteredClaims
	DiscordUsername string `json:"discord_username"`
}
