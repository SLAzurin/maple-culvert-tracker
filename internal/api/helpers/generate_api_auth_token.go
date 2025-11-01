package helpers

import (
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/slazurin/maple-culvert-tracker/internal/data"
)
/*
GenerateAPIAuthToken generates a JWT token for API authentication

discordMemberID: usually `i.Member.User.Username` or use `s.State.User.Username + "#" + s.State.User.Discriminator`
*/
func GenerateAPIAuthToken(discordMemberID string, expiry time.Time) string {
	mode := 0 // not boolean because json takes 4 character width instead of 1 character width
	if os.Getenv(gin.EnvGinMode) != gin.ReleaseMode {
		mode = 1
	}
	claims := &data.MCTClaims{
		DiscordUsername: discordMemberID,
		DiscordServerID: os.Getenv(data.EnvVarDiscordGuildID),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiry),
		},
		DevMode: mode,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(os.Getenv(data.EnvVarJWTSecret)))

	return tokenString
}
