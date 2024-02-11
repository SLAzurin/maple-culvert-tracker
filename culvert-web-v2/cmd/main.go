package main

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

type BaseHtmxProps struct {
	Title string
}

var AuthenticatedMiddleware gin.HandlerFunc = func(c *gin.Context) {
	host := "localhost"
	if os.Getenv("GIN_MODE") == "release" {
		host = os.Getenv("FRONTEND_URL")
	}
	auth, err := c.Cookie("mct_token")
	if err != nil || auth == "" {
		c.SetCookie("mct_token", "", -1, "/", host, true, true)
		c.Redirect(http.StatusTemporaryRedirect, "/login")
		c.Abort()
		return
	}
	c.Set("mct_token", auth)

}

func main() {
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

	cp := r.Group("/control-panel")
	{
		cp.Use(AuthenticatedMiddleware)
		cp.GET("/", func(ctx *gin.Context) {
			controlpanelindex().Render(ctx, ctx.Writer)
		})
	}

	r.Run(":8080")
}
