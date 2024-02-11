package main

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
	"github.com/slazurin/maple-culvert-tracker/internal/data"
)

var AuthenticatedMiddleware gin.HandlerFunc = func(c *gin.Context) {
	godotenv.Load("../.env")

	host := "localhost"
	if os.Getenv("GIN_MODE") == "release" {
		host = os.Getenv("FRONTEND_URL")
	}
	auth, err := c.Cookie("mct_token")
	if err != nil || auth == "" {
		c.Header("HX-Location", "/login")
		c.SetCookie("mct_token", "", -1, "/", host, true, true)
		c.Redirect(http.StatusTemporaryRedirect, "/login")
		c.Abort()
		return
	}
	c.Set("mct_token", auth)
	c.Next()
}

func main() {
	// Use modd and nodemon -e .go --exec "go run ./cmd" --signal SIGTERM for dev
	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusPermanentRedirect, "/login")
		c.Abort()
	})

	r.GET("/login", func(c *gin.Context) {
		auth, _ := c.Cookie("mct_token")
		if auth != "" {
			c.Redirect(http.StatusTemporaryRedirect, "/control-panel")
			c.Abort()
			return
		}
		login().Render(c, c.Writer)
	})

	CPAPI(r)

	cp := r.Group("/control-panel")
	{
		cp.Use(AuthenticatedMiddleware)
		cp.GET("/", func(ctx *gin.Context) {
			controlpanelindex().Render(ctx, ctx.Writer)
		})
	}

	r.Run(":8080")
}

func CPAPI(r *gin.Engine) {
	cpapi := r.Group("/cpapi")
	{
		cpapi.POST("/validate-auth", func(c *gin.Context) {
			auth := c.Request.Header.Get("Authorization")
			if strings.HasPrefix(auth, "Bearer ") {
				auth = auth[7:]
			} else {
				c.AbortWithStatus(http.StatusUnauthorized)
				return
			}
			claims := &data.MCTClaims{}
			tkn, err := jwt.ParseWithClaims(auth, claims, func(token *jwt.Token) (interface{}, error) {
				return []byte(os.Getenv("JWT_SECRET")), nil
			})
			if err != nil {
				if err == jwt.ErrSignatureInvalid {
					c.AbortWithStatus(http.StatusUnauthorized)
					return
				}
				c.AbortWithStatus(http.StatusUnauthorized)
				return
			}
			if !tkn.Valid {
				c.AbortWithStatus(http.StatusUnauthorized)
				return
			}
			c.AbortWithStatus(http.StatusOK)
		})
	}
}
