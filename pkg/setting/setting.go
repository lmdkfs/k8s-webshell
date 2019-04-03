package setting

import (
	"os"
	"strings"

	"github.com/spf13/viper"
)

var (
	RunMode  string
	HTTPPort string

	LogPath string
	LogName string

	KubeConfig        string
	InCluster         bool
	SslCertificate    string
	SslCertificateKey string
	JwtSecret         string

	SecretKey string

)

type Config struct {
	vp *viper.Viper
}

type Configer interface {
	LoadConfig() error
	LoadServer()
}

func LoadBaseConfig(conf Configer) {
	conf.LoadConfig()
}

func init() {
	var config Config
	LoadBaseConfig(&config)
	config.LoadServer()

}

func (config *Config) LoadConfig() error {
	config.vp = viper.New()
	// load config from env prefix harbortools
	config.vp.SetEnvPrefix("webshell")                         // 环境变量前缀,环境变量必须大写.
	config.vp.SetEnvKeyReplacer(strings.NewReplacer(".", "_")) // 为了兼容读取yaml文件
	config.vp.AutomaticEnv()

	// load config from yaml
	config.vp.AddConfigPath("/Users/zhuruiqing/go/src/k8s-webshell/configs")
	config.vp.AddConfigPath("/Users/finup/GoglandProjects/src/k8s-webshell/configs")
	config.vp.AddConfigPath("./")
	if os.Getenv("env") == "production" {
		config.vp.SetConfigName("config")

	} else {
		config.vp.SetConfigName("devconfig")
	}
	//fmt.Println(os.Getenv("env"))
	//utils.Logger.Info("Current Env: ", os.Getenv("env"))

	config.vp.SetConfigType("yaml")
	if err := config.vp.ReadInConfig(); err != nil {
		return err
	}

	return nil
}

func (config *Config) LoadServer() {

	RunMode = config.vp.GetString("server.run_mode")
	HTTPPort = config.vp.GetString("server.port")
	LogPath = config.vp.GetString("server.logpath")
	LogName = config.vp.GetString("server.logname")
	KubeConfig = config.vp.GetString("server.kubeconfig")
	InCluster = config.vp.GetBool("server.incluster")
	SslCertificate = config.vp.GetString("server.ssl_certificate")
	SslCertificateKey = config.vp.GetString("server.ssl_certificate_key")
	JwtSecret = config.vp.GetString("server.jwt_secret")
	SecretKey = config.vp.GetString("server.secret_key")


}
