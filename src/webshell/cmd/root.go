package cmd

import (
	"fmt"
	"github.com/mitchellh/go-homedir"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"log"
	"os"
	"strings"
	"webshell/webshell/common"
	"webshell/webshell/config"
	"webshell/webshell/server"
)
var cfgFile string
var cfg = config.NewConfig()

var rootCmd = &cobra.Command{
	Use: "k8s-webshell",
	Short: "k8s-webshell",
	Long: "k8s-webshell",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Start server")
		server.Start()
	},

}

// Execute ...
func Execute() {
	if err := rootCmd.Execute(); err !=nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.cfg.yaml)")
	rootCmd.PersistentFlags().StringP("port","p", "8888","listen port")
	//rootCmd.PersistentFlags().StringP("host","h", "0.0.0.0","listen port")
	rootCmd.PersistentFlags().String("kubeconf", "","kubeconfig file")
	rootCmd.PersistentFlags().Bool("incluster", true,"incluster mode")
	rootCmd.PersistentFlags().String("ssl-crt","", "ssl_certificate")
	rootCmd.PersistentFlags().String("ssl-key","", "ssl_certificate")
	rootCmd.PersistentFlags().String("jwt-secret","", "jwt_secret")
	rootCmd.PersistentFlags().String("secret-key","", "secret_key")
	rootCmd.PersistentFlags().String("logpath","", "logpath")
	rootCmd.PersistentFlags().StringP("logpname","l","k8s-webshell.log", "logpath")

	viper.BindPFlag("server.port", rootCmd.PersistentFlags().Lookup("port"))
	//viper.BindPFlag("server.host", rootCmd.PersistentFlags().Lookup("host"))
	viper.BindPFlag("k8s.kubeconfig", rootCmd.PersistentFlags().Lookup("kubeconf"))
	viper.BindPFlag("k8s.incluster", rootCmd.PersistentFlags().Lookup("incluster"))
	viper.BindPFlag("server.certificate", rootCmd.PersistentFlags().Lookup("ssl-crt"))
	viper.BindPFlag("server.certificatekey", rootCmd.PersistentFlags().Lookup("ssl-key"))
	viper.BindPFlag("server.jwtsecret", rootCmd.PersistentFlags().Lookup("jwt-secret"))
	viper.BindPFlag("server.secretkey", rootCmd.PersistentFlags().Lookup("secret-key"))
	viper.BindPFlag("server.logpath", rootCmd.PersistentFlags().Lookup("logpath"))
	viper.BindPFlag("server.logpname",rootCmd.PersistentFlags().Lookup("logpname"))


}

func initConfig() {

	currentDir, err := os.Getwd()
	if err != nil {
		log.Panicf("Get currentir Fail: %s ", err.Error())
	}
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := homedir.Dir()
		if err != nil {
			log.Printf("Get $HOME Dir error:%s", err.Error())
			os.Exit(1)
		}
		viper.AddConfigPath(currentDir)
		viper.AddConfigPath(home)
		viper.SetConfigName(".cfg")
	}


	// Search config in home directory with name ".config" (without extension)

	viper.SetEnvPrefix("webshell") // 环境变量前缀,环境变量必须大写.
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_")) // 为了兼容读取yaml文
	viper.AutomaticEnv()

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		log.Println("Using config file:", viper.ConfigFileUsed())

	}

	cfg.RunMode = viper.GetString("server.run_mode")
	cfg.HTTP.Host = viper.GetString("server.host")
	cfg.HTTP.Port = viper.GetInt("server.port")
	cfg.HTTP.Certificate = viper.GetString("server.certificate")
	cfg.HTTP.CertificateKey = viper.GetString("server.certificatekey")

	cfg.Log.LogName = viper.GetString("server.logname")
	cfg.Log.LogPath = viper.GetString("server.Logpath")


	cfg.K8s.KubeConfig = viper.GetString("k8s.kubeconfig")
	cfg.K8s.InCluster = viper.GetBool("k8s.incluster")

	cfg.Security.JWTSecret = viper.GetString("server.jwtsecret")
	cfg.Security.SecretKey = viper.GetString("server.secretkey")


	exist, err := PathExists(cfg.Log.LogPath)
	logrus.Println("日志路径:", cfg.Log.LogPath)
	if err != nil {
		logrus.Errorf("Get dir error![%v]\n", err)
	}
	if !exist {
		err := os.MkdirAll(cfg.Log.LogPath, 0755)
		if err != nil {
			logrus.Errorf("MkDirs Failed![%v]\n", err)
		} else {
			logrus.Info("MkDirs Success!\n")
		}

	}
	fmt.Println("Incluster status:", cfg.K8s.InCluster)
	err = common.InitClient()
	if err != nil {
		logrus.Panicf("init kube-config fail:%s", err.Error())
	}
}

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}