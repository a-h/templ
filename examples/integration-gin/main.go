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

	r.Run(":8080")
}
