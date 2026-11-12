package helpers

import (
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/slazurin/maple-culvert-tracker/internal/data"
)

const AuthCookieName = "mct_token"

func ParseAuthToken(tokenString string) (*data.MCTClaims, error) {
	claims := &data.MCTClaims{}
	tkn, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
		return []byte(os.Getenv(data.EnvVarJWTSecret)), nil
	})
	if err != nil {
		return nil, err
	}
	if !tkn.Valid {
		return nil, jwt.ErrTokenInvalidClaims
	}
	return claims, nil
}

func SetAuthCookie(c *gin.Context, tokenString string, expiresAt time.Time) {
	maxAge := int(time.Until(expiresAt).Seconds())
	if maxAge < 0 {
		maxAge = 0
	}
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(AuthCookieName, tokenString, maxAge, "/", "", true, true)
}

func ClearAuthCookie(c *gin.Context) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(AuthCookieName, "", -1, "/", "", true, true)
}
