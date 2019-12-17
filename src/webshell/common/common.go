package common

import (
	"io/ioutil"
	"sync"

	"webshell/webshell/utils"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	once      sync.Once
	clientSet *kubernetes.Clientset
	err       error
)

func GetK8sCli() (*kubernetes.Clientset) {
	return clientSet

}

func InitClient() (err error) {
	var (
		restConf *rest.Config
	)
	if restConf, err = GetRestConf(); err != nil {
		utils.Logger.Info("GetRestConfig Client error:", err)
		return
	}

	if clientSet, err = kubernetes.NewForConfig(restConf); err != nil {
		utils.Logger.Info("init k8s client error:", err)
		goto END
	}

END:
	return
}

func GetRestConf() (restConf *rest.Config, err error) {
	var (
		kubeconfig []byte
	)

	switch cfg.K8s.InCluster {
	case true:
		utils.Logger.Info("Start run in cluster mode")
		if restConf, err = rest.InClusterConfig(); err != nil {
			goto END
		}
	case false:
		//rest.InClusterConfig()
		// 读取kubeconfig文件

		//utils.Logger.Info("kubeconfig file ", setting.KubeConfig)
		if kubeconfig, err = ioutil.ReadFile(cfg.K8s.KubeConfig); err != nil {
			goto END
		}

		if restConf, err = clientcmd.RESTConfigFromKubeConfig(kubeconfig); err != nil {
			goto END
		}

	}

END:
	return
}
