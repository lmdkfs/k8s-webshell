package config

import (
	"sync"
	"time"
)

var (
	global *Config
	once   sync.Once
)

func NewConfig() *Config {
	once.Do(func() {
		global = &Config{}
	})
	return global
}

type Config struct {
	RunMode  string
	HTTP     HTTP
	K8s      K8s
	Security Security
	Log      Log
}

type HTTP struct {
	Host            string
	Port            int
	ShutdownTimeout int

	Certificate    string
	CertificateKey string
	SSLEnable      bool
	ReadTimeout    time.Duration
	WriteTimeout   time.Duration
}

type K8s struct {
	KubeConfig string
	InCluster  bool
}

type Security struct {
	JWTSecret string
	SecretKey string
}

type Log struct {
	LogPath string
	LogName string
	Level   int
	Format  string
}
