package middlewares

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/slazurin/maple-culvert-tracker/internal/api/helpers"
)

func extractToken(c *gin.Context) (string, bool) {
	if cookie, err := c.Cookie(helpers.AuthCookieName); err == nil && cookie != "" {
		return cookie, true
	}
	auth := c.Request.Header.Get("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		return auth[7:], true
	}
	return "", false
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, ok := extractToken(c)
		if !ok {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		claims, err := helpers.ParseAuthToken(tokenString)
		if err != nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		c.Set("claims", claims)
		c.Set("discord_username", claims.DiscordUsername)
		c.Set("discord_server_id", claims.DiscordServerID)
		c.Set("discord_user_id", claims.DiscordUserID)

		c.Next()
	}
}
