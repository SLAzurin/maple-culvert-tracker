package controllers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/slazurin/maple-culvert-tracker/internal/api/helpers"
	"github.com/slazurin/maple-culvert-tracker/internal/data"
)

type AuthController struct{}

type loginBody struct {
	Token string `json:"token" binding:"required"`
}

type claimsResponse struct {
	Exp             int64  `json:"exp"`
	DiscordUsername string `json:"discord_username"`
	DiscordServerID string `json:"discord_server_id"`
	DiscordUserID   string `json:"discord_user_id"`
	DevMode         int    `json:"dev_mode"`
}

func claimsToResponse(claims *data.MCTClaims) claimsResponse {
	var exp int64
	if claims.ExpiresAt != nil {
		exp = claims.ExpiresAt.Unix()
	}
	return claimsResponse{
		Exp:             exp,
		DiscordUsername: claims.DiscordUsername,
		DiscordServerID: claims.DiscordServerID,
		DiscordUserID:   claims.DiscordUserID,
		DevMode:         claims.DevMode,
	}
}

func (AuthController) Login(c *gin.Context) {
	var body loginBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.AbortWithStatus(http.StatusBadRequest)
		return
	}

	claims, err := helpers.ParseAuthToken(body.Token)
	if err != nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	expiresAt := time.Now().Add(4 * time.Hour)
	if claims.ExpiresAt != nil {
		expiresAt = claims.ExpiresAt.Time
	}
	helpers.SetAuthCookie(c, body.Token, expiresAt)
	c.JSON(http.StatusOK, claimsToResponse(claims))
}

func (AuthController) Logout(c *gin.Context) {
	helpers.ClearAuthCookie(c)
	c.Status(http.StatusNoContent)
}

func (AuthController) Me(c *gin.Context) {
	claims, ok := c.Get("claims")
	if !ok {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	mctClaims, ok := claims.(*data.MCTClaims)
	if !ok {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	c.JSON(http.StatusOK, claimsToResponse(mctClaims))
}
