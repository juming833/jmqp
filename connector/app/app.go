package app

import (
	"common/config"
	"common/logs"
	"connector/route"
	"context"
	"core/repo"
	"framework/connector"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func Run(ctx context.Context, serverId string) error {
	//日志
	logs.InitLog(config.Conf.AppName)
	exit := func() {}
	go func() {
		c := connector.Default()
		exit = c.Close
		manager := repo.New()
		c.RegisterHandler(route.RegisterHandler(manager))
		c.Run(serverId)
	}()

	stop := func() {
		exit()
		time.Sleep(3 * time.Second)
		logs.Info("stop app finish")
	}
	//优雅启停
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT, syscall.SIGHUP)
	for {
		select {
		case <-ctx.Done():
			stop()
			return nil
		case s := <-c:
			switch s {
			case syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGINT:
				stop()
				logs.Info("connector app quit")
				return nil
			case syscall.SIGHUP:
				stop()
				logs.Info("connector app reload")
				return nil
			default:
				return nil
			}

		}

	}
}
