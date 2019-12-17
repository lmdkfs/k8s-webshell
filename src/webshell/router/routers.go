package router

import (
	"time"

	"webshell/webshell/config"
	"webshell/webshell/controllers"
	"webshell/webshell/middlewares/jwt"
	"webshell/webshell/utils"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func InitRouter() *gin.Engine {

	cfg := config.NewConfig()
	route := gin.New()

	gin.SetMode(cfg.RunMode)
	route.Use(utils.GinRus(utils.Logger, time.RFC3339, false))
	route.Use(gin.Recovery())
	route.POST("/auth", controllers.GetAuth)
	route.GET("/metrics", gin.WrapH(promhttp.Handler()))
	apiV1 := route.Group("/api")
	apiV1.Use(jwt.JWT())
	{

		apiV1.GET("/ws", controllers.WsHandler)
		apiV1.POST("/ws", controllers.WsHandler)

	}
	return route
}
