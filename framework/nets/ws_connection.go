package nets

import (
	"common/logs"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"sync/atomic"
	"time"
)

var CidBase uint64 = 10000
var (
	maxMessageSize int64 = 1024
	PongWait             = 10 * time.Second
	writeWait            = 10 * time.Second
	pingInterval         = (PongWait * 9) / 10
)

type WsConnection struct {
	Cid        string //客户端id
	Conn       *websocket.Conn
	manager    *Manager
	ReadChan   chan *MsgPack
	WriteChan  chan []byte
	Session    *Session
	pingTicker *time.Ticker
}

func NewWsConnection(conn *websocket.Conn, manager *Manager) *WsConnection {
	cid := fmt.Sprintf("%s-%s-%d", uuid.New().String(), manager.ServerId, atomic.AddUint64(&CidBase, 1))
	return &WsConnection{
		Conn:      conn,
		manager:   manager,
		Cid:       cid,
		WriteChan: make(chan []byte, 1024),
		ReadChan:  manager.ClientReadChan,
		Session:   NewSession(cid),
	}

}
func (c *WsConnection) Run() {
	go c.readMessage()
	go c.writeMessage()
	//心跳检测websocket的ping，pong机制
	c.Conn.SetPongHandler(c.PongHandler)

}
func (c *WsConnection) PongHandler(data string) error {
	//logs.Info("pong........")
	if err := c.Conn.SetReadDeadline(time.Now().Add(PongWait)); err != nil {
		return err

	}
	return nil
}

func (c *WsConnection) Close() {
	if c.Conn != nil {
		c.Conn.Close()
	}
	if c.pingTicker != nil {
		c.pingTicker.Stop()
	}
}
func (c *WsConnection) GetSession() *Session {
	return c.Session
}

func (c *WsConnection) readMessage() {
	defer func() {
		c.manager.removeClient(c)
	}()
	c.Conn.SetReadLimit(maxMessageSize)
	if err := c.Conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
		logs.Error("SetWriteDeadline  err:%v", err)
		return
	}
	for {
		messageType, message, err := c.Conn.ReadMessage()
		if err != nil {
			break
		}
		//客户端发来的消息为二进制消息
		if messageType == websocket.BinaryMessage {
			if c.ReadChan != nil {
				c.ReadChan <- &MsgPack{
					Cid:  c.Cid,
					Body: message,
				}
			}
		} else {
			logs.Error("不支持此消息类型:%d", messageType)
		}
	}

}
func (c *WsConnection) writeMessage() {
	//if c.pingTicker != nil {
	//	c.pingTicker.Stop()
	//}
	ticker := time.NewTicker(pingInterval)
	for {
		select {
		case message, ok := <-c.WriteChan:
			if !ok {
				if err := c.Conn.WriteMessage(websocket.CloseMessage, nil); err != nil {
					logs.Error("connection closed,%v", err)
				}
				return
			}
			if err := c.Conn.WriteMessage(websocket.BinaryMessage, message); err != nil {
				logs.Error("client[%s] write message err:%v", c.Cid, err)
			}
		case <-ticker.C:
			if err := c.Conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				logs.Error("client[%s] ping SetWriteDeadline  err:%v", c.Cid, err)
			}
			//logs.Info("ping........")
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				logs.Error("client[%s] ping message err:%v", c.Cid, err)
			}
		}
	}

}
func (c *WsConnection) SendMessage(buf []byte) error {
	c.WriteChan <- buf
	return nil
}
