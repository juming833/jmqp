package nets

import (
	"common/logs"
	"common/utils"
	"encoding/json"
	"errors"
	"fmt"
	"framework/game"
	"framework/protocol"
	"framework/remote"
	"github.com/gorilla/websocket"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"
)

var (
	websocketUpgrade = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

type CheckOriginHandler func(r *http.Request) bool
type Manager struct {
	sync.RWMutex
	ServerId          string
	websocketUpgrade  *websocket.Upgrader
	CheckOriginHandle CheckOriginHandler
	clients           map[string]Connection
	ClientReadChan    chan *MsgPack
	handlers          map[protocol.PackageType]EventHandler
	ConnectorHandlers LogicHandler
	RemoteReadChan    chan []byte
	RemoteCli         remote.Client
	RemotePushChan    chan *remote.Msg
}
type HandleFunc func(session *Session, body []byte) (any, error)
type LogicHandler map[string]HandleFunc
type EventHandler func(packet *protocol.Packet, c Connection) error

func NewManager() *Manager {
	return &Manager{
		ClientReadChan: make(chan *MsgPack, 1024),
		clients:        make(map[string]Connection),
		handlers:       make(map[protocol.PackageType]EventHandler),
		RemoteReadChan: make(chan []byte, 1024),
		RemotePushChan: make(chan *remote.Msg, 1024),
	}
}
func (m *Manager) Run(addr string) {

	go m.ClientReadChanHandler()
	go m.RemoteReadChanHandler()
	go m.RemotePushChanHandler()

	http.HandleFunc("/", m.serveWS)

	m.setupEventHandlers()
	logs.Fatal("connector listen serve err :%v", http.ListenAndServe(addr, nil))

}

func (m *Manager) serveWS(w http.ResponseWriter, r *http.Request) {

	if m.websocketUpgrade == nil {
		m.websocketUpgrade = &websocketUpgrade
	}
	wsConn, err := m.websocketUpgrade.Upgrade(w, r, nil)
	if err != nil {
		logs.Error("websocket upgrade err: %v", err)
		return

	}
	client := NewWsConnection(wsConn, m)
	m.addClient(client)
	client.Run()
}
func (m *Manager) addClient(client *WsConnection) {
	m.Lock()
	defer m.Unlock()
	m.clients[client.Cid] = client
}

func (m *Manager) removeClient(wc *WsConnection) {
	for cid, c := range m.clients {
		if cid == wc.Cid {
			c.Close()
			delete(m.clients, cid)
		}
	}
}
func (m *Manager) ClientReadChanHandler() {
	for {
		select {
		case body, ok := <-m.ClientReadChan:
			if ok {
				m.decodeClientPack(body)
			}
		}
	}
}
func (m *Manager) decodeClientPack(body *MsgPack) {
	//解析协议
	//logs.Info("receive message:%v", string(body.Body))
	packet, err := protocol.Decode(body.Body)
	if err != nil {
		logs.Error("decode message err: %v", err)
		return
	}
	if err := m.routeEvent(packet, body.Cid); err != nil {

		logs.Error("routeEvent err11111111111: %v", err)
	}
}

func (m *Manager) Close() {
	for cid, v := range m.clients {
		v.Close()
		delete(m.clients, cid)
	}

}
func (m *Manager) routeEvent(packet *protocol.Packet, cid string) error {
	//根据packet.type做不同处理
	conn, ok := m.clients[cid]
	if ok {
		handler, ok := m.handlers[packet.Type]
		if ok {
			return handler(packet, conn)
		} else {
			return errors.New("not found packetType")
		}
	}
	return errors.New("not found client ")

}
func (m *Manager) setupEventHandlers() {

	m.handlers[protocol.Handshake] = m.HandshakeHandler
	m.handlers[protocol.HandshakeAck] = m.HandshakeAckHandler
	m.handlers[protocol.Heartbeat] = m.HeartbeatHandler
	m.handlers[protocol.Data] = m.MessageHandler
	m.handlers[protocol.Kick] = m.KickHandler

}
func (m *Manager) HandshakeHandler(packet *protocol.Packet, c Connection) error {
	res := protocol.HandshakeResponse{
		Code: 200,
		Sys: protocol.Sys{
			Heartbeat: 3,
		},
	}
	data, err := json.Marshal(res)
	buf, err := protocol.Encode(packet.Type, data)
	if err != nil {
		logs.Error("encode packet err: %v", err)
		return err
	}

	return c.SendMessage(buf)
}
func (m *Manager) HandshakeAckHandler(packet *protocol.Packet, c Connection) error {
	logs.Info("receiver Hands hakeAckHandler message >>>>> ")
	return nil
}
func (m *Manager) HeartbeatHandler(packet *protocol.Packet, c Connection) error {
	logs.Info("receiver HeartbeatHandler message :%v ", packet.Type)
	var res []byte
	data, err := json.Marshal(res)
	buf, err := protocol.Encode(packet.Type, data)
	if err != nil {
		logs.Error("encode packet err: %v", err)
		return err
	}

	return c.SendMessage(buf)

}
func (m *Manager) MessageHandler(packet *protocol.Packet, c Connection) error {

	message := packet.MessageBody()
	logs.Info("receiver MessageHandler message, type=%v router=%v,data:%v ", message.Type, message.Route, string(message.Data))
	routeStr := message.Route
	routers := strings.Split(routeStr, ".")
	if len(routers) != 3 {
		return errors.New("route err")
	}
	serverType := routers[0]
	HandleMethod := fmt.Sprintf("%s.%s", routers[1], routers[2])
	connectorConfig := game.Conf.GetConnectorByServerType(serverType)
	if connectorConfig != nil {
		handle, ok := m.ConnectorHandlers[HandleMethod]
		if ok {
			data, err := handle(c.GetSession(), message.Data)
			if err != nil {
				return err
			}
			marshal, _ := json.Marshal(data)
			message.Type = protocol.Response
			message.Data = marshal
			encode, err := protocol.MessageEncode(message)
			if err != nil {
				return err
			}
			res, err := protocol.Encode(packet.Type, encode)
			if err != nil {
			}
			return c.SendMessage(res)
		}
	} else {
		//nats远端调用处理 hall.userHandler.updateUserAddress

		dst, err := m.selectDst(serverType)
		if err != nil {
			logs.Error("selectDst err: %v", err)
			return err
		}
		msg := &remote.Msg{
			Cid:         c.GetSession().Cid,
			Uid:         c.GetSession().Uid,
			Src:         m.ServerId,
			Dst:         dst,
			Router:      HandleMethod,
			Body:        message,
			SessionData: c.GetSession().data,
		}
		data, _ := json.Marshal(msg)
		err = m.RemoteCli.SendMsg(dst, data)
		if err != nil {
			logs.Error("remote send msg err: %v", err)
			return err
		}

	}
	return nil

}
func (m *Manager) KickHandler(packet *protocol.Packet, c Connection) error {
	logs.Info("receiver KickHandler message >>>>> ")

	return nil
}
func (m *Manager) RemoteReadChanHandler() {
	for {
		select {
		case body, ok := <-m.RemoteReadChan:
			if ok {
				logs.Info("sub nats msg :%v", string(body))
				var msg remote.Msg
				err := json.Unmarshal(body, &msg)
				if err != nil {
					logs.Error("nats remote message format err: %v", err)
					continue
				}
				if msg.Type == remote.SessionType {
					m.setSessionData(msg)

					continue

				}
				if msg.Body != nil {
					if msg.Body.Type == protocol.Request || msg.Body.Type == protocol.Response {
						msg.Body.Type = protocol.Response
						m.Response(&msg)
					}
					if msg.Body.Type == protocol.Push {
						m.RemotePushChan <- &msg
					}

				}
			}
		}
	}

}
func (m *Manager) selectDst(serverType string) (string, error) {
	serversConfigs, ok := game.Conf.ServersConf.TypeServer[serverType]
	if !ok {
		return "", errors.New("not found serverType")
	}

	//随机
	rand.New(rand.NewSource(time.Now().UnixNano()))
	index := rand.Intn(len(serversConfigs))
	return serversConfigs[index].ID, nil

}

func (m *Manager) Response(msg *remote.Msg) {
	connection, ok := m.clients[msg.Cid]
	if !ok {
		logs.Info("%s client  not found,uid=%s", msg.Cid, msg.Uid)
		return
	}
	buf, err := protocol.MessageEncode(msg.Body)
	if err != nil {
		logs.Error(" response encode message err: %v", err)
		return
	}
	res, err := protocol.Encode(protocol.Data, buf)
	if err != nil {
		logs.Error(" message encode err: %v", err)

		return
	}
	if msg.Body.Type == protocol.Push {

		for _, v := range m.clients {

			if utils.Contains(msg.PushUser, v.GetSession().Uid) {
				v.SendMessage(res)
				return

			}

		}
	}
	connection.SendMessage(res)

}

func (m *Manager) RemotePushChanHandler() {
	for {
		select {
		case body, ok := <-m.RemotePushChan:
			if ok {
				if body.Body.Type == protocol.Push {
					m.Response(body)
				}
			}
		}
	}

}

func (m *Manager) setSessionData(msg remote.Msg) {
	m.RLock()
	defer m.RUnlock()
	connection, ok := m.clients[msg.Cid]
	if ok {
		connection.GetSession().SetData(msg.Uid, msg.SessionData)
	}

}
