package datebase

import (
	"common/config"
	"common/logs"
	"context"
	"github.com/redis/go-redis/v9"
	"time"
)

type RedisManager struct {
	ClusterCli *redis.ClusterClient //集群
	Cli        *redis.Client        //单机
}

func NewRedis() *RedisManager {
	var clusterCli *redis.ClusterClient
	var cli *redis.Client
	clusterAddrs := config.Conf.Database.RedisConf.ClusterAddrs
	if len(clusterAddrs) <= 0 {
		//单节点
		cli = redis.NewClient(&redis.Options{
			Addr:         config.Conf.Database.RedisConf.Addr,
			PoolSize:     config.Conf.Database.RedisConf.PoolSize,
			MinIdleConns: config.Conf.Database.RedisConf.MinIdleConns,
			Password:     config.Conf.Database.RedisConf.Password,
		})
	} else {
		//集群
		clusterCli = redis.NewClusterClient(&redis.ClusterOptions{
			Addrs:        config.Conf.Database.RedisConf.ClusterAddrs,
			PoolSize:     config.Conf.Database.RedisConf.PoolSize,
			MinIdleConns: config.Conf.Database.RedisConf.MinIdleConns,
			Password:     config.Conf.Database.RedisConf.Password,
		})
	}
	if clusterCli != nil {
		err := clusterCli.Ping(context.TODO()).Err()
		if err != nil {
			logs.Fatal("redis cluster ping err: %v", err)
		}
	} else {
		err := cli.Ping(context.TODO()).Err()
		if err != nil {
			logs.Fatal("redis ping err: %v", err)
		}
	}
	return &RedisManager{
		Cli:        cli,
		ClusterCli: clusterCli,
	}
}
func (r *RedisManager) Close() {
	if r.ClusterCli != nil {
		if err := r.ClusterCli.Close(); err != nil {
			logs.Error("redis cluster close err: %v", err)
		}
	}
	if r.Cli != nil {
		if err := r.Cli.Close(); err != nil {
			logs.Error("redis  close err: %v", err)
		}
	}
}

func (r *RedisManager) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	if r.ClusterCli != nil {
		return r.ClusterCli.Set(ctx, key, value, expiration).Err()
	}
	if r.Cli != nil {
		return r.Cli.Set(ctx, key, value, expiration).Err()
	}
	return nil

}
