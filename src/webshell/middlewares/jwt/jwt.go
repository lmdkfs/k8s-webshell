package jwt

import (
	"net/http"
	"webshell/webshell/common"
	"webshell/webshell/common/e"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

func JWT() gin.HandlerFunc {
	return func(c *gin.Context) {
		var code int
		var data interface{}

		code = e.SUCCESS
		token := c.Query("token")

		if token == "" {

			code = e.INVALID_PARAMS
		} else {
			claims, err := common.ParseToken(token)

			if err != nil {

				switch err.(*jwt.ValidationError).Errors {

				case jwt.ValidationErrorExpired:
					code = e.ERROR_AUTH_CHECK_TOKEN_TIMEOUT
				default:
					code = e.ERROR_AUTH_CHECK_TOKEN_FAIL

				}

			} else {
				c.Set("podNs", claims.PodNs)
				c.Set("podName", claims.PodName)
				c.Set("containerName", claims.ContainerName)
				c.Set("paasUser", claims.PaasUser)

			}

		}
		if code != e.SUCCESS {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code": code,
				"msg":  e.GetMsg(code),
				"data": data,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
