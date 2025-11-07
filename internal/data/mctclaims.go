package data

import (
	"github.com/golang-jwt/jwt/v5"
)

type MCTClaims struct {
	jwt.RegisteredClaims
	DiscordUsername string `json:"discord_username"`
	DiscordServerID string `json:"discord_server_id"`
	DiscordUserID   string `json:"discord_user_id"`
	DevMode         int    `json:"dev_mode"`
}
