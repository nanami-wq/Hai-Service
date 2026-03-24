package gin

import (
	"Hai-Service/api/middleware"
	"Hai-Service/api/router"
	"Hai-Service/config"
	"github.com/gin-gonic/gin"
)

func GinInit() *gin.Engine {
	r := gin.Default()
	config.MustLoad()
	//api init
	router.GenerateRouter(r)
	middleware.InitSecret(config.GetConfig().JWT.Secret)

	return r
}
