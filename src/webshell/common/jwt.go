package common

import (
	"time"

	"webshell/webshell/config"

	"github.com/dgrijalva/jwt-go"
)

var cfg = config.NewConfig()
var jwtSecret = []byte(cfg.Security.JWTSecret)

type Claims struct {
	SecretKey     string `json:"secretkey"`
	PaasUser      string `json:"paasuser"`
	PodNs         string `json:"podNs"`
	PodName       string `json:"podName"`
	ContainerName string `json:"containerName"`

	jwt.StandardClaims
}

func GenerateToken(secretKey, paasUser, podNs, podName, containerName string) (string, error) {
	nowTime := time.Now()
	expireTime := nowTime.Add(20 * time.Minute)

	claims := Claims{
		secretKey,
		paasUser,
		podNs,
		podName,
		containerName,
		jwt.StandardClaims{
			ExpiresAt: expireTime.Unix(),
			Issuer:    "k8s-webshell",
		},
	}
	tokenClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, err := tokenClaims.SignedString(jwtSecret)
	return token, err
}

func ParseToken(token string) (*Claims, error) {

	tokenClaims, err := jwt.ParseWithClaims(token, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if tokenClaims != nil {
		if Claims, ok := tokenClaims.Claims.(*Claims); ok && tokenClaims.Valid {
			return Claims, nil
		}
	}
	return nil, err
}
