package common

import (
	"io/ioutil"

	"k8s-webshell/pkg/setting"
	"k8s-webshell/pkg/utils"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

func InitClient() (clientset *kubernetes.Clientset, err error) {
	var (
		restConf *rest.Config
	)
	if restConf, err = GetRestConf(); err != nil {
		utils.Logger.Info("GetRestConfig Client error:", err)
		return
	}

	if clientset, err = kubernetes.NewForConfig(restConf); err != nil {
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

	switch setting.InCluster {
	case true:
		utils.Logger.Info("Start run in cluster mode")
		if restConf, err = rest.InClusterConfig(); err != nil {
			goto END
		}
	case false:
		//rest.InClusterConfig()
		// 读取kubeconfig文件

		//utils.Logger.Info("kubeconfig file ", setting.KubeConfig)
		if kubeconfig, err = ioutil.ReadFile(setting.KubeConfig); err != nil {
			goto END
		}

		if restConf, err = clientcmd.RESTConfigFromKubeConfig(kubeconfig); err != nil {
			goto END
		}

	}

END:
	return
}
