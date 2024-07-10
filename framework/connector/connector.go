package connector

import (
	"common/logs"
	"fmt"
	"framework/game"
	"framework/nets"
	"framework/remote"
)

type Connector struct {
	isRunning bool
	wsManager *nets.Manager
	handles   nets.LogicHandler
	remoteCli remote.Client
}

func Default() *Connector {
	return &Connector{
		wsManager: nets.NewManager(),
		handles:   make(nets.LogicHandler),
		//handles: make(map[string]nets.HandleFunc),
	}

}
func (c *Connector) Run(ServerId string) {
	if !c.isRunning {
		//启动websocket和nats
		c.wsManager = nets.NewManager()
		c.wsManager.ConnectorHandlers = c.handles
		//启动nats
		c.remoteCli = remote.NewNatClient(ServerId, c.wsManager.RemoteReadChan)
		c.remoteCli.Run()
		c.wsManager.RemoteCli = c.remoteCli
		c.Serve(ServerId)
	}

}
func (c *Connector) Close() {
	if c.wsManager != nil {
		//关闭websocket和nats
		c.wsManager.Close()
	}

}
func (c *Connector) Serve(serverId string) {
	logs.Info("run connector server id :%v", serverId)
	//地址，需要读取配置文件
	c.wsManager.ServerId = serverId
	connectorConfig := game.Conf.GetConnector(serverId)
	if connectorConfig == nil {
		logs.Fatal("no connector config found")
	}
	addr := fmt.Sprintf("%s:%d", connectorConfig.Host, connectorConfig.ClientPort)
	c.isRunning = true
	c.wsManager.Run(addr)
}
func (c *Connector) RegisterHandler(handles nets.LogicHandler) {
	c.handles = handles
}
