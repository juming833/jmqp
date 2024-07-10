package main

import (
	"common/config"
	"common/metrics"
	"context"
	"fmt"
	"framework/game"
	"github.com/spf13/cobra"
	"hall/app"
	"log"
	"os"
)

//func main() {
//	//连接websocket
//	//1.wsmanager2.natsclient
//	c := connector.Default()
//	c.Run()
//
//}

var rootCmd = &cobra.Command{
	Use:   "hall",
	Short: "hall 大厅相关处理",
	Long:  `connector 管理连接，session以及路由请求`,
	Run: func(cmd *cobra.Command, args []string) {
	},
	PostRun: func(cmd *cobra.Command, args []string) {
	},
}

// var configFile = flag.String("config", "application.yml", "config file")
var (
	configFile    string
	gameConfigDir string
	serverId      string
)

func init() {
	rootCmd.Flags().StringVar(&configFile, "config", "application.yml", "app config yml file")
	rootCmd.Flags().StringVar(&gameConfigDir, "gameDir", "../config", "game config dir")
	rootCmd.Flags().StringVar(&serverId, "serverId", "", "app server id， required")
	_ = rootCmd.MarkFlagRequired("serverId")
}
func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Println(err)
		os.Exit(1)
	}
	config.InitConfig(configFile)

	game.InitConfig(gameConfigDir)

	go func() {
		err := metrics.Serve(fmt.Sprintf("0.0.0.0:%d", config.Conf.MetricPort))
		if err != nil {
			panic(err)
		}
	}()
	err := app.Run(context.Background(), serverId)
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
}
