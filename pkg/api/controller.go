package api

import (
	"bytes"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"k8s-webshell/pkg/common"
	"k8s-webshell/pkg/utils"
	"k8s-webshell/pkg/ws"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/json"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
	"regexp"
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
	logBuff     *bytes.Buffer
	moveCursor  int
}

// web终端发来的包
type xtermMessage struct {
	MsgType string `json:"type"`  // 类型: resize 客户端调整终端,input客户端输入
	Input   string `json:"input"` // msgtype=input 情况下使用
	Rows    uint16 `json:"rows"`  // msgtype=resize情况下使用
	Cols    uint16 `json:"cols"`  // msgtype=resize情况下使用
}

// executor 回调获取web是否resize
func (handler *streamHandler) Next() (size *remotecommand.TerminalSize) {
	ret := <-handler.resizeEvent
	size = &ret
	return
}

func (handler *streamHandler) RuneSliceDeleteStr() {

	defer func() {
		if r := recover(); r != nil {
			utils.Logger.Warn("Recovered in RuneSliceDeleteStr:", r)
		}
	}()
	runeSlice := []rune(handler.logBuff.String())

	if len(runeSlice) > handler.moveCursor {
		deleteIndex := len(runeSlice) - handler.moveCursor
		runeSlice = append(runeSlice[:deleteIndex-1], runeSlice[deleteIndex:]...)
		handler.logBuff.Reset()
		handler.logBuff.WriteString(string(runeSlice))
	}

	runeSlice = nil
}
func (handler *streamHandler) RuneSliceInsertStr(insertString *string) {
	defer func() {
		if r := recover(); r != nil {
			utils.Logger.Warn("Recovered in RuneSliceInsertStr:", r)
		}
	}()
	runeSlice := []rune(handler.logBuff.String())
	insertIndex := len(runeSlice) - handler.moveCursor
	runeSlice = append(runeSlice[:insertIndex], append([]rune(*insertString), runeSlice[insertIndex:]...)...)
	handler.logBuff.Reset()
	handler.logBuff.WriteString(string(runeSlice))

	runeSlice = nil

}

func (handler *streamHandler) RecordCommand(inputString *string) {
	defer func() {
		if r := recover(); r != nil {
			utils.Logger.Warn("Recovered in RecordCommand:", r)
		}
	}()
	var (
		n int
	)

	if len(*inputString) > 0 {
		n = len(*inputString) - 1
	}
	invalidChart, _ := regexp.MatchString(`\s?\[\d{1,100}\;9R`, *inputString)
	leftMoveCursor, _ := regexp.MatchString(`\s?\[D`, *inputString)
	rightMoveCursor, _ := regexp.MatchString(`\s?\[C`, *inputString)
	switch {
	case invalidChart:
		utils.Logger.Info("enter >> :", []rune(*inputString))

	case leftMoveCursor:
		cmdLens := len([]rune(handler.logBuff.String()))
		if cmdLens-handler.moveCursor != 0 {
			handler.moveCursor += 1
		}

	case rightMoveCursor:
		handler.moveCursor -= 1

	case []byte(*inputString)[n] == 12: //12  FF  (NP form feed, new page)
		utils.Logger.WithFields(logrus.Fields{
			"PassUser":  handler.paasUser,
			"PodName":   handler.podName,
			"NameSpace": handler.podNs,
			"serviceName":"k8s-webshell",
			"command":   "clear screen",
		}).Info("record input")

	case []byte(*inputString)[n] == 13: // 13  CR  (carriage return)
		handler.moveCursor = 0          // cursor flag reset
		if len(*inputString) > 1 {
			handler.logBuff.WriteString(*inputString)
		}
		if len(handler.logBuff.String()) > 0 {
			utils.Logger.WithFields(logrus.Fields{
				"PassUser":    handler.paasUser,
				"PodName":     handler.podName,
				"NameSpace":   handler.podNs,
				"serviceName": "k8s-webshell",
				"command":     handler.logBuff.String(),
			}).Info("record input")
		}

		handler.logBuff.Reset()

	//case []byte(*inputString)[n] == 37:
	//	utils.Logger.Info("fangxiangjian", []byte(*inputString)[n])
	case []byte(*inputString)[n] == 127: // 127  DEL

		if len([]rune(handler.logBuff.String())) > 0 {
			handler.RuneSliceDeleteStr()
		}

	case []byte(*inputString)[n] == 3:
		utils.Logger.WithFields(logrus.Fields{
			"PassUser":    handler.paasUser,
			"PodName":     handler.podName,
			"NameSpace":   handler.podNs,
			"serviceName": "k8s-webshell",
			"command":     "ctrl + c",
		}).Info("record input")
		handler.logBuff.Reset()

	default:
		if handler.moveCursor != 0 {
			handler.RuneSliceInsertStr(inputString)
		} else {
			handler.logBuff.WriteString(*inputString)
		}

	}

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

	if err = json.Unmarshal(msg.Data, &xtermMsg); err != nil {
		return
	}
	// web ssh 调整了终端大小
	switch xtermMsg.MsgType {
	case "resize":
		// 放到channel里, 等remotecommand executor 调用我们的Next取走
		handler.resizeEvent <- remotecommand.TerminalSize{Width: xtermMsg.Cols, Height: xtermMsg.Rows}
	case "input":
		// web ssh 终端输入了字符
		size = len(xtermMsg.Input)
		// copy 到p数组中
		copy(p, xtermMsg.Input)
		handler.RecordCommand(&xtermMsg.Input)
	}

	return

}

// executor 回调想web 输出
func (handler *streamHandler) Write(p []byte) (size int, err error) {
	size = len(p)
	copy := append(make([]byte, 0, size), p...) // 解决 发送数据丢失的问题

	err = handler.wsConn.WsWrite(websocket.BinaryMessage, copy)
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

	podNs = c.MustGet("podNs").(string)
	podName = c.MustGet("podName").(string)
	containerName = c.MustGet("containerName").(string)
	paasUser = c.MustGet("paasUser").(string)
	utils.Logger.Infof("nameSpaces:%s, podName: %s, containerName:%s", podNs, podName, containerName)

	// 得到websocket 长连接
	if wsConn, err = ws.InitWebsocket(c.Writer, c.Request); err != nil {
		utils.Logger.Info("up to ws error:", err)
		return
	}
	//var logBuff bytes.Buffer

	logBuff := bytes.NewBufferString("")
	// 获取k8s rest client 配置
	if restConf, err = common.GetRestConf(); err != nil {
		utils.Logger.Info("get kubeconfig error ", err)
		goto END
	}
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
	// 创建到容器的连接
	if executor, err = remotecommand.NewSPDYExecutor(restConf, "POST", sshReq.URL()); err != nil {
		utils.Logger.Info("创建到容器的连接失败:", err)
		goto END
	}

	// 配置与容器之间的数据流处理回调

	handler = &streamHandler{
		wsConn:      wsConn,
		resizeEvent: make(chan remotecommand.TerminalSize),
		podName:     &podName,
		podNs:       &podNs,
		paasUser:    &paasUser,
		logBuff:     logBuff,
		moveCursor:  0,
	}
	utils.Logger.Infof("Start to exec command from pod:%s,", podName)
	if err = executor.Stream(remotecommand.StreamOptions{
		Stdin:             handler,
		Stdout:            handler,
		Stderr:            handler,
		TerminalSizeQueue: handler,
		Tty:               true,
	}); err != nil {
		goto END
	}

	defer handler.logBuff.Reset()

	return

END:
	utils.Logger.Info("exec command error: ", err)
	wsConn.WsClose()

	return
}
