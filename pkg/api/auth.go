package api

import (
	"github.com/gin-gonic/gin"
	"k8s-webshell/pkg/e"
	"k8s-webshell/pkg/setting"
	"k8s-webshell/pkg/utils"
	"net/http"
)

type auth struct {
	Username string
	Password string
}


func GetAuth(c *gin.Context) {
	username := c.Query("username")
	password := c.Query("password")
	podNs := c.Query("podNs")
	podName := c.Query("podName")
	containerName := c.Query("containerName")
	data := make(map[string]interface{})
	code := e.INVALID_PARAMS

	if username == setting.UserName && password == setting.PassWord {
		token, err := utils.GenerateToken(username, password, podNs, podName, containerName)
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
