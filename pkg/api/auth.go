package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"k8s-webshell/pkg/e"
	"k8s-webshell/pkg/setting"
	"k8s-webshell/pkg/utils"
	"net/http"
)

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
	fmt.Println(">>>", apiAuth.SecretKey,
		apiAuth.SecretKey,
		apiAuth.PaasUser,
		apiAuth.PodNs,
		apiAuth.PodName,
		apiAuth.ContainerName)

	if apiAuth.SecretKey == setting.SecretKey {
		token, err := utils.GenerateToken(
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
		utils.Logger.Info("username or password invalid")
	}

	c.JSON(http.StatusOK, gin.H{
		"code": code,
		"msg":  e.GetMsg(code),
		"data": data,
	})
}
