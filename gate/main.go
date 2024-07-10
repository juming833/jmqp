package main

import (
	"common/config"
	"common/metrics"
	"context"
	"flag"
	"fmt"
	"gate/app"
	"log"
	"os"
)

var configFile = flag.String("config", "application.yml", "config file")

func main() {
	flag.Parse()
	config.InitConfig(*configFile)
	//启动监听
	go func() {
		err := metrics.Serve(fmt.Sprintf("0.0.0.0:%d", config.Conf.MetricPort))
		if err != nil {
			panic(err)
		}
	}()
	err := app.Run(context.Background())
	if err != nil {
		log.Println(err)
		os.Exit(-1)
	}
}
