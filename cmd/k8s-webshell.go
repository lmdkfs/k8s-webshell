package main

import (
	"os"

	"k8s-webshell/pkg/setting"
	"k8s-webshell/pkg/utils"
	"k8s-webshell/routers"
)

func main() {
	ginServer := routers.InitRouter()
	//ginpprof.Wrapper(ginServer)
	utils.Logger.Info("Current ENV: ", os.Getenv("env"))
	utils.Logger.Info("Start k8s-webshell on Port: ", setting.HTTPPort)
	err := ginServer.RunTLS(":"+setting.HTTPPort, setting.SslCertificate, setting.SslCertificateKey)

	if err != nil {
		utils.Logger.Fatal("Gin  Start err", err)
	}

}
