package controllers

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"webshell/webshell/common"
	"webshell/webshell/common/e"
	"webshell/webshell/config"
	"webshell/webshell/utils"
)

var cfg = config.NewConfig()

type apiAuthInfo struct {
	SecretKey     string `from:"secretKey" binding:"required"`
	PaasUser      string `from:"paasUser" binding:"required"`
	PodNs         string `from:"rpodNs" binding:"required"`
	PodName       string `from:"podName" binding:"required"`
	ContainerName string `from:"containerName" binding:"required"`
}

func GetAuth(c *gin.Context) {

	var apiAuth apiAuthInfo
	data := make(map[string]interface{})
	code := e.INVALID_PARAMS

	if c.Bind(&apiAuth) != nil {
		utils.Logger.Info("解析json失败")

		c.JSON(http.StatusBadRequest, gin.H{
			"code": code,
			"msg":  e.GetMsg(code),
			"data": data,
		})
		return
	}

	if apiAuth.SecretKey == cfg.Security.SecretKey {
		token, err := common.GenerateToken(
			apiAuth.SecretKey,
			apiAuth.PaasUser,
			apiAuth.PodNs,
			apiAuth.PodName,
			apiAuth.ContainerName)
		if err != nil {
			code = e.ERROR_AUTH_TOKEN
		} else {
			data["token"] = token
			code = e.SUCCESS
		}
	} else {
		code = e.ERROR_AUTH_CHECK_TOKEN_FAIL
		utils.Logger.Info("username or password invalid")
	}

	c.JSON(http.StatusOK, gin.H{
		"code": code,
		"msg":  e.GetMsg(code),
		"data": data,
	})
}
