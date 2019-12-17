package server

import (
	"os"
	"fmt"
	"webshell/webshell/config"
	"webshell/webshell/router"
	"webshell/webshell/utils"
)

func Start() {
	cfg := config.NewConfig()
	fmt.Println("",cfg.Log.LogPath)
	ginServer := router.InitRouter()
	//ginpprof.Wrapper(ginServer)
	utils.Logger.Info("Current ENV: ", os.Getenv("env"))
	utils.Logger.Info("Start k8s-webshell on Port: ", cfg.HTTP.Port)
	//err := ginServer.RunTLS(":"+ string(cfg.HTTP.Port), cfg.HTTP.Certificate, cfg.HTTP.CertificateKey)
	err := ginServer.RunTLS(":"+ "7777", cfg.HTTP.Certificate, cfg.HTTP.CertificateKey)

	if err != nil {
		utils.Logger.Fatalf("Gin  Start err: %s", err.Error())
	}
}
