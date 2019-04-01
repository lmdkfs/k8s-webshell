package jwt

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"k8s-webshell/pkg/e"
	"k8s-webshell/pkg/utils"
	"net/http"
	"time"
)

func JWT() gin.HandlerFunc {
	return func(c *gin.Context) {
		var code int
		var data interface{}

		code = e.SUCCESS
		token := c.Query("token")

		if token == "" {
			fmt.Println("token is empty")
			code = e.INVALID_PARAMS
		} else {
			claims, err := utils.ParseToken(token)
			if err != nil {
				code = e.ERROR_AUTH_CHECK_TOKEN_FAIL
			} else if time.Now().Unix() > claims.ExpiresAt {
				code = e.ERROR_AUTH_CHECK_TOKEN_TIMEOUT
			}
		}

		if code != e.SUCCESS {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":code,
				"msg": e.GetMsg(code),
				"data": data,
			})
			c.Abort()
			return
		}
		utils.Logger.Info("jwt next")
		fmt.Println("next>>>>")
		c.Next()
	}
}