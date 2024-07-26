package main

import (
	"common/config"
	"common/metrics"
	"context"
	"fmt"
	"framework/game"
	"game/app"
	"github.com/spf13/cobra"
	"log"
	"os"
)

var rootCmd = &cobra.Command{
	Use:   "game1",
	Short: "游戏逻辑相关处理",
	Long:  `game游戏逻辑相关处理 `,
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
	rootCmd.Flags().StringVar(&gameConfigDir, "gameDir", "../config", "game1 config dir")
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
