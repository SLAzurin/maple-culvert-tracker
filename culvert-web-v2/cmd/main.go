package main

import (
	"github.com/a-h/templ"
	"github.com/gin-gonic/gin"
)

type BaseHtmxProps struct {
	Title         string
	BodyComponent func() templ.Component
}

func main() {
	r := gin.Default()
	// http.Handle("/", templ.Handler(hello("Hello world")))
	r.GET("/", func(c *gin.Context) {
		basehtmx(BaseHtmxProps{
			Title:         "Maple Culvert Tracker",
			BodyComponent: func() templ.Component { return login() },
		}).Render(c, c.Writer)
	})

	r.Run(":8080")
}
