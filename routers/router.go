package routers

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"

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
	route.POST("/auth", api.GetAuth)
	route.GET("/metrics", gin.WrapH(promhttp.Handler()))
	apiV1 := route.Group("/api")
	apiV1.Use(jwt.JWT())
	{

		apiV1.GET("/ws", api.WsHandler)
		apiV1.POST("/ws", api.WsHandler)

	}

	return route
}
