package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	r.HTMLRender = &TemplRender{}

	r.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "", Home())
	})

	r.GET("/with-ctx", func(c *gin.Context) {
		c.Render(http.StatusOK, &TemplRender{
			Code: http.StatusOK,
			Data: Home(),
			Ctx:  c.Request.Context(),
		})
	})

	r.Run(":8080")
}
