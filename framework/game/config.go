package game

import (
	"common/logs"
	"encoding/json"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"io"
	"log"
	"os"
	"path"
)

var Conf *Config

const (
	gameConfig = "gameConfig.json"
	servers    = "servers.json"
)

type Config struct {
	GameConfig  map[string]GameConfigValue `json:"gameConfig"`
	ServersConf ServersConf                `json:"serversConf"`
}
type ServersConf struct {
	Nats       NatsConfig         `json:"nats"`
	Connector  []*ConnectorConfig `json:"connector"`
	Servers    []*ServersConfig   `json:"servers"`
	TypeServer map[string][]*ServersConfig
}

type ServersConfig struct {
	ID               string `json:"id"`
	ServerType       string `json:"serverType"`
	HandleTimeOut    int    `json:"handleTimeOut"`
	RPCTimeOut       int    `json:"rpcTimeOut"`
	MaxRunRoutineNum int    `json:"maxRunRoutineNum"`
}

type ConnectorConfig struct {
	ID         string `json:"id"`
	Host       string `json:"host"`
	ClientPort int    `json:"clientPort"`
	Frontend   bool   `json:"frontend"`
	ServerType string `json:"serverType"`
}
type NatsConfig struct {
	Url string `json:"url"`
}
type GameConfigValue map[string]any

func InitConfig(configDir string) {
	Conf = new(Config)
	dir, err := os.ReadDir(configDir)
	if err != nil {
		logs.Fatal("read config dir err:%v", err)
	}
	for _, v := range dir {
		configFile := path.Join(configDir, v.Name())
		if v.Name() == gameConfig {
			readGameConfig(configFile)

		}
		if v.Name() == servers {
			readServersConfig(configFile)
		}
	}
}
func readGameConfig(configFile string) {
	//var gameConfig = make(map[string]GameConfigValue)
	//v := viper.New()
	//v.SetConfigFile(configFile)
	//v.WatchConfig()
	//v.OnConfigChange(func(e fsnotify.Event) {
	//	log.Println("gameConfig配置文件被修改")
	//	err := v.Unmarshal(&gameConfig)
	//	if err != nil {
	//		panic(fmt.Errorf("gameConfi配置文件被修改以后，报错，err:%v \n", err))
	//	}
	//	Conf.GameConfig = gameConfig
	//
	//})
	//err := v.ReadInConfig()
	//if err != nil {
	//	panic(fmt.Errorf("gameConfi读取配置文件报错，err:%v \n", err))
	//}
	//if err := v.Unmarshal(&gameConfig); err != nil {
	//	panic(fmt.Errorf("gameConfi Unmarshal data to Conf failed ，err:%v \n", err))
	//}
	//Conf.GameConfig = gameConfig
	file, err := os.Open(configFile)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	data, err := io.ReadAll(file)
	if err != nil {
		panic(err)
	}
	var gameConfig = make(map[string]GameConfigValue)
	err = json.Unmarshal(data, &gameConfig)
	if err != nil {
		panic(err)
	}
	Conf.GameConfig = gameConfig
}
func readServersConfig(configFile string) {
	var serversConf ServersConf
	v := viper.New()
	v.SetConfigFile(configFile)
	v.WatchConfig()
	v.OnConfigChange(func(e fsnotify.Event) {
		log.Println("serversConf配置文件被修改")
		err := v.Unmarshal(&serversConf)
		if err != nil {
			panic(fmt.Errorf("serversConf配置文件被修改以后，报错，err:%v \n", err))
		}
		Conf.ServersConf = serversConf
		typeServerConfig()
	})
	err := v.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("读取serversConf配置文件报错，err:%v \n", err))
	}
	if err := v.Unmarshal(&serversConf); err != nil {
		panic(fmt.Errorf("Unmarshal data to Conf failed ，err:%v \n", err))
	}
	Conf.ServersConf = serversConf
	typeServerConfig()
}

func typeServerConfig() {
	if len(Conf.ServersConf.Servers) > 0 {
		if Conf.ServersConf.TypeServer == nil {
			Conf.ServersConf.TypeServer = make(map[string][]*ServersConfig)
		}
		for _, v := range Conf.ServersConf.Servers {
			if Conf.ServersConf.TypeServer[v.ServerType] == nil {
				Conf.ServersConf.TypeServer[v.ServerType] = make([]*ServersConfig, 0, 10)
			}
			Conf.ServersConf.TypeServer[v.ServerType] = append(Conf.ServersConf.TypeServer[v.ServerType], v)
		}
	}
}
func (c *Config) GetConnector(serverId string) *ConnectorConfig {

	for _, v := range c.ServersConf.Connector {
		if v.ID == serverId {
			return v
		}
	}
	return nil
}
func (c *Config) GetConnectorByServerType(serverType string) *ConnectorConfig {

	for _, v := range c.ServersConf.Connector {
		if v.ServerType == serverType {
			return v
		}
	}
	return nil
}
func (c *Config) GetFrontGameConfig() map[string]any {
	result := make(map[string]any)
	for k, v := range c.GameConfig {
		value, ok := v["value"]
		backend := false
		_, exist := v["backend"]
		if exist {
			backend = v["backend"].(bool)
		}
		if ok && !backend {
			result[k] = value
		}
	}
	return result
}
