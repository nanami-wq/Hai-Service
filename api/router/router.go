package router

import (
	"Hai-Service/api/controller"
	"Hai-Service/api/middleware"
	"github.com/gin-gonic/gin"
)

func GenerateRouter(r *gin.Engine) {
	r.Use(gin.Recovery(), middleware.CorsWare())
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	g1 := r.Group("/api/sv", middleware.JWTAuthMiddleware())
	pic := controller.NewPictureController()
	pic.Register(g1)
}
