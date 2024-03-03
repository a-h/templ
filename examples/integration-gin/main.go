package main

import (
	"net/http"

	"github.com/a-h/templ/examples/integration-gin/gintemplrenderer"
	"github.com/gin-gonic/gin"
)

func main() {
	engine := gin.Default()
	engine.LoadHTMLFiles("./home.html")

	//engine.HTMLRender = gintemplrenderer.Default

	ginHtmlRenderer := engine.HTMLRender
	engine.HTMLRender = &gintemplrenderer.HTMLTemplRenderer{FallbackHtmlRenderer: ginHtmlRenderer}

	// Disable trusted proxy warning.
	engine.SetTrustedProxies(nil)

	engine.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "", Home())
	})

	engine.GET("/with-ctx", func(c *gin.Context) {
		r := gintemplrenderer.New(c.Request.Context(), http.StatusOK, Home())
		c.Render(http.StatusOK, r)
	})

	engine.GET("/with-fallback-renderer", func(c *gin.Context) {
		c.HTML(http.StatusOK, "home.html", gin.H{})
	})

	engine.Run(":8080")
}
