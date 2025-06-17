package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/samber/do"
)

func RegisterRoutes(server *gin.Engine, injector *do.Injector) {
	server.GET("/", func(c *gin.Context) {
		c.File("user_management_frontend.html")
	})
	User(server, injector)
	DocumentRoutes(server, injector)
	SignatureRoutes(server, injector)
}
