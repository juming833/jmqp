package app

import (
	"common/config"
	"common/discovery"
	"common/logs"
	"context"
	"core/repo"
	"google.golang.org/grpc"
	_ "google.golang.org/grpc/reflection"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
	"user/internal/service"
	"user/pb"
)

func Run(ctx context.Context) error {
	//日志
	logs.InitLog(config.Conf.AppName)

	//etcd
	register := discovery.NewRegister()
	//启用grpc服务端
	server := grpc.NewServer()
	//初始化数据库
	manager := repo.New()
	go func() {
		lis, err := net.Listen("tcp", config.Conf.Grpc.Addr)
		if err != nil {
			logs.Fatal("grpc监听错误:%v", err)
		}
		//注册grpc service
		err = register.Register(config.Conf.Etcd)

		if err != nil {
			logs.Fatal("etcd register err:%v", err)
		}
		pb.RegisterUserServiceServer(server, service.NewAccountService(manager))
		//阻塞操作
		err = server.Serve(lis)

		if err != nil {
			logs.Fatal("grpc监服务错误:%v", err)
		}
	}()

	stop := func() {
		server.Stop()
		register.Close()
		manager.Close()
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
				logs.Info("user app quit")
				return nil
			case syscall.SIGHUP:
				stop()
				logs.Info("user app reload")
				return nil
			default:
				return nil
			}

		}

	}
}
