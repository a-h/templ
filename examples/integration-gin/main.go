package main

import (
	"net/http"

	"github.com/a-h/templ/examples/integration-gin/gintemplrenderer"
	"github.com/gin-gonic/gin"
)

func main() {
	engine := gin.Default()
	engine.HTMLRender = gintemplrenderer.Default

	// Disable trusted proxy warning.
	engine.SetTrustedProxies(nil)

	engine.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "", Home())
	})

	engine.GET("/with-ctx", func(c *gin.Context) {
		r := gintemplrenderer.New(c.Request.Context(), http.StatusOK, Home())
		c.Render(http.StatusOK, r)
	})

	engine.Run(":8080")
}
