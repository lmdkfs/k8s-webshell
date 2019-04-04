package ws

import (
	"k8s-webshell/pkg/utils"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
)

var wsUpgrader = websocket.Upgrader{
	// 允许所有CORS跨域访问
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// websocket 消息
type WsMessage struct {
	MessageType int
	Data        []byte
}

//封装websocket 连接

type WsConnection struct {
	wsSocket *websocket.Conn // 底层websocket
	inChan   chan *WsMessage // 读取队列
	outChan  chan *WsMessage // 发送队列

	mutex     sync.Mutex // 避免重复关闭管道
	isClosed  bool
	closeChan chan byte // 关闭通知
}

// 读协程
func (wsConn *WsConnection) wsReadLoop() {
	var (
		msgType int
		data    []byte
		msg     *WsMessage
		err     error
	)

	for {
		if msgType, data, err = wsConn.wsSocket.ReadMessage(); err != nil {
			utils.Logger.Info("读取协程错误:", err)
			goto ERROR
		}

		msg = &WsMessage{
			MessageType: msgType,
			Data:        data,
		}

		select {
		case wsConn.inChan <- msg:
		case <-wsConn.closeChan:
			goto CLOSED
		}

	}
ERROR:
	wsConn.WsClose()

CLOSED:
}

// 发送协程

func (wsConn *WsConnection) wsWriteLoop() {
	var (
		msg *WsMessage
		err error
	)

	for {
		select {
		case msg = <-wsConn.outChan:
			if err = wsConn.wsSocket.WriteMessage(msg.MessageType, msg.Data); err != nil {
				utils.Logger.Info("发送错误:", err)
				goto ERROR
			}
		case <-wsConn.closeChan:
			goto CLOSED
		}
	}
ERROR:
	wsConn.WsClose()
CLOSED:
}

/******并发安全api***/

func InitWebsocket(resp http.ResponseWriter, req *http.Request) (wsConn *WsConnection, err error) {
	var (
		wsSocket *websocket.Conn
	)

	// 应答客户端告知升级连接为websocket
	if wsSocket, err = wsUpgrader.Upgrade(resp, req, nil); err != nil {
		utils.Logger.Info("update to ws error:", err)
		return
	}

	wsConn = &WsConnection{
		wsSocket:  wsSocket,
		inChan:    make(chan *WsMessage, 1000),
		outChan:   make(chan *WsMessage, 1000),
		closeChan: make(chan byte),
		isClosed:  false,
	}
	// 读协程
	go wsConn.wsReadLoop()
	// 写协程

	go wsConn.wsWriteLoop()

	return
}

//发送消息
func (wsConn *WsConnection) WsWrite(messageType int, data []byte) (err error) {
	select {
	case wsConn.outChan <- &WsMessage{messageType, data,}:
	case <-wsConn.closeChan:
		err = errors.New("websocket closed")
	}
	return
}

// 读取消息

func (wsConn *WsConnection) WsRead() (msg *WsMessage, err error) {
	select {
	case msg = <-wsConn.inChan:
		return
	case <-wsConn.closeChan:
		err = errors.New("websocket closed")
	}

	return
}

// 关闭连接

func (wsConn *WsConnection) WsClose() {
	wsConn.wsSocket.Close()
	wsConn.mutex.Lock()
	defer wsConn.mutex.Unlock()
	if !wsConn.isClosed {
		wsConn.isClosed = true
		close(wsConn.closeChan)
	}
}
