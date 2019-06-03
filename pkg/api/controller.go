package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"k8s-webshell/pkg/common"
	"k8s-webshell/pkg/utils"
	"k8s-webshell/pkg/ws"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

var (
	ClientSet *kubernetes.Clientset
	err       error
)

func init() {

	if ClientSet, err = common.InitClient(); err != nil {
		utils.Logger.Panic("init k8s client err", err)

	}

}

// ssh 流式处理器
type streamHandler struct {
	wsConn      *ws.WsConnection
	resizeEvent chan remotecommand.TerminalSize
	podName     *string
	podNs       *string
	paasUser    *string
}

// web终端发来的包
type xtermMessage struct {
	MsgType string `json:"type"`  // 类型: resize 客户端调整终端,input客户端输入
	Input   string `json:"input"` // msgtype=input 情况下使用
	Rows    uint16 `json:"rows"`  // msgtype=resize情况下使用
	Cols    uint16 `json:"cols"`  // msgtype=resize情况下使用
}

//type PodInfo struct {
//	podNs         string `form:"podNs"`
//	podName       string `form:"podName"`
//	containerName string `form:"containerName"`
//}

// executor 回调获取web是否resize
func (handler *streamHandler) Next() (size *remotecommand.TerminalSize) {
	ret := <-handler.resizeEvent
	size = &ret
	return
}

// executor 回调读取web端的输入

func (handler *streamHandler) Read(p []byte) (size int, err error) {
	var (
		msg      *ws.WsMessage
		xtermMsg xtermMessage
	)
	// 读web发来的输入

	if msg, err = handler.wsConn.WsRead(); err != nil {
		return
	}
	// 解析客户端请求
	if err = json.Unmarshal(msg.Data, &xtermMsg); err != nil {
		return
	}
	// web ssh 调整了终端大小
	if xtermMsg.MsgType == "resize" {
		// 放到channel里, 等remotecommand executor 调用我们的Next取走
		handler.resizeEvent <- remotecommand.TerminalSize{Width: xtermMsg.Cols, Height: xtermMsg.Rows}
	} else if xtermMsg.MsgType == "input" { // web ssh 终端输入了字符
		// copy 到p数组中
		size = len(xtermMsg.Input)
		//utils.Logger.Info("webShell Input:", xtermMsg.Input)
		copy(p, xtermMsg.Input)

	}
	return

}

// executor 回调想web 输出
func (handler *streamHandler) Write(p []byte) (size int, err error) {
	size = len(p)
	//err = handler.wsConn.WsWrite(websocket.TextMessage, p)
	fmt.Println("send to webterminal:", string(p), len(p))
	err = handler.wsConn.WsWrite(websocket.BinaryMessage, p)
	return
}
func WsHandler(c *gin.Context) {
	var (
		wsConn        *ws.WsConnection
		restConf      *rest.Config
		sshReq        *rest.Request
		podName       string
		podNs         string
		containerName string
		executor      remotecommand.Executor
		handler       *streamHandler
		err           error
		paasUser      string
	)

	// 解析GET 参数

	//podNs = c.Query("podNs")
	//podName = c.Query("podName")
	//containerName = c.Query("containerName")
	podNs = c.MustGet("podNs").(string)
	podName = c.MustGet("podName").(string)
	containerName = c.MustGet("containerName").(string)
	paasUser = c.MustGet("paasUser").(string)
	utils.Logger.Infof("nameSpaces:%s, podName: %s, containerName:%s", podNs, podName, containerName)
	//fmt.Println(">>>>>", podNs, podName, containerName)

	//utils.Logger.Info("get kwargs  **")
	// 得到websocket 长连接
	if wsConn, err = ws.InitWebsocket(c.Writer, c.Request); err != nil {
		utils.Logger.Info("up to ws error:", err)
		return
	}

	//podName = "my-nginx-f9995bdb6-bb8sr"
	//podNs = "default"
	//containerName = "my-nginx"

	// 获取k8s rest client 配置
	if restConf, err = common.GetRestConf(); err != nil {
		utils.Logger.Info("get kubeconfig error ", err)
		goto END
	}
	//utils.Logger.Info("start to k8s post", podName, podNs, containerName)
	sshReq = ClientSet.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(podName).
		Namespace(podNs).
		SubResource("exec").
		VersionedParams(&v1.PodExecOptions{
			Container: containerName,
			Command:   []string{"sh"},
			Stdin:     true,
			Stdout:    true,
			Stderr:    true,
			TTY:       true,
		}, scheme.ParameterCodec)
	//utils.Logger.Info("end k8s post")
	// 创建到容器的连接
	if executor, err = remotecommand.NewSPDYExecutor(restConf, "POST", sshReq.URL()); err != nil {
		utils.Logger.Info("创建到容器的连接失败:", err)
		goto END
	}

	// 配置与容器之间的数据流处理回调
	handler = &streamHandler{wsConn: wsConn, resizeEvent: make(chan remotecommand.TerminalSize), podName: &podName, podNs: &podNs, paasUser: &paasUser}
	if err = executor.Stream(remotecommand.StreamOptions{
		Stdin:             handler,
		Stdout:            handler,
		Stderr:            handler,
		TerminalSizeQueue: handler,
		Tty:               true,
	}); err != nil {
		goto END
	}
	return

END:
	utils.Logger.Info("exec command error: ", err)
	wsConn.WsClose()

	return
}
