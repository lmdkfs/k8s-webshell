package routers

import (
	"time"

	"github.com/gin-gonic/gin"
	"k8s-webshell/middleware/jwt"
	"k8s-webshell/pkg/api"
	"k8s-webshell/pkg/setting"
	"k8s-webshell/pkg/utils"
)

func InitRouter() *gin.Engine {

	route := gin.New()

	gin.SetMode(setting.RunMode)
	route.Use(utils.GinRus(utils.Logger, time.RFC3339, false))
	route.Use(gin.Recovery())
	route.GET("/auth", api.GetAuth)
	apiV1 := route.Group("/api")
	apiV1.Use(jwt.JWT())
	{

		apiV1.GET("/ws", api.WsHandler)

	}

	return route
}
