package routers

import (
	"time"

	"k8s-webshell/pkg/api"
	"k8s-webshell/pkg/setting"
	"k8s-webshell/pkg/utils"

	"github.com/gin-gonic/gin"
)

func InitRouter() *gin.Engine {

	route := gin.New()

	gin.SetMode(setting.RunMode)
	route.Use(utils.GinRus(utils.Logger, time.RFC3339, false))
	route.Use(gin.Recovery())
	apiV1 := route.Group("/api")
	{

		apiV1.GET("/ws", api.WsHandler)

	}

	return route
}
