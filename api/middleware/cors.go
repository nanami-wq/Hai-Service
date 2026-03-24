package middleware

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// CorsWare 自定义 CORS 中间件
func CorsWare() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent) // 处理 OPTIONS 预检请求
			return
		}

		c.Next()
	}
}
