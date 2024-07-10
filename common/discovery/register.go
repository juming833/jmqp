package discovery

import (
	"common/config"
	"common/logs"
	"context"
	"encoding/json"
	clientv3 "go.etcd.io/etcd/client/v3"
	"time"
)

// Register 将grpc注册到etcd
// 原理 创建一个租约 将grpc服务信息注册到etcd并且绑定租约
// 如果过了租约时间，etcd会删除存储的信息
// 可以实现心跳，完成续租，如果etcd没有则重新注册

type Register struct {
	etcdCli     *clientv3.Client                        //etcd连接
	leaseId     clientv3.LeaseID                        //租约id
	DialTimeout int                                     //超时时间
	ttl         int64                                   //租约时间
	keepAliveCh <-chan *clientv3.LeaseKeepAliveResponse //心跳channel
	info        Server                                  //注册的服务信息
	closeCh     chan struct{}
}

func NewRegister() *Register {
	return &Register{
		DialTimeout: 3,
	}
}
func (r *Register) Close() {
	r.closeCh <- struct{}{}

}
func (r *Register) Register(conf config.EtcdConf) error {
	//注册信息
	info := Server{
		Name:    conf.Register.Name,
		Addr:    conf.Register.Addr,
		Version: conf.Register.Version,
		Weight:  conf.Register.Weight,
		Ttl:     conf.Register.Ttl,
	}
	//建立连接
	var err error
	r.etcdCli, err = clientv3.New(clientv3.Config{
		Endpoints:   conf.Addrs,
		DialTimeout: time.Duration(r.DialTimeout) * time.Second,
	})
	if err != nil {
		return err
	}
	r.info = info
	if err := r.register(); err != nil {
		return err
	}

	r.closeCh = make(chan struct{})
	go r.watcher()
	return nil

}
func (r *Register) register() error {
	//创建租约
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(r.DialTimeout)*time.Second)
	defer cancel()
	err := r.createLease(ctx, r.info.Ttl)
	if err != nil {
		return err
	}
	//心跳检测
	r.keepAliveCh, err = r.keepAlive()
	if err != nil {
		return err
	}

	//绑定租约
	data, err := json.Marshal(r.info)
	if err != nil {
		logs.Error("marshal register info failed,err:%v", err)
	}

	return r.bindLease(ctx, r.info.BuildRegisterKey(), string(data))
}
func (r *Register) bindLease(ctx context.Context, key string, value string) error {
	_, err := r.etcdCli.Put(ctx, key, value, clientv3.WithLease(r.leaseId))
	if err != nil {
		logs.Error("bind lease err:%v", err)
		return err
	}
	logs.Info("register service success,key=%s", key)
	return nil
}

func (r *Register) createLease(ctx context.Context, ttl int64) error {
	grant, err := r.etcdCli.Grant(ctx, ttl)
	if err != nil {
		logs.Error("create lease err:%v", err)
		return err
	}
	r.leaseId = grant.ID
	return nil
}

func (r *Register) keepAlive() (<-chan *clientv3.LeaseKeepAliveResponse, error) {
	keepAliveResponses, err := r.etcdCli.KeepAlive(context.Background(), r.leaseId)
	if err != nil {
		logs.Error("keep alive err:%v", err)
		return keepAliveResponses, err
	}
	return keepAliveResponses, nil
}

func (r *Register) watcher() {
	ticker := time.NewTicker(time.Duration(r.info.Ttl) * time.Second)
	for {
		select {
		case <-r.closeCh:
			if err := r.unregister(); err != nil {
				logs.Error("unregister err:%v", err)
			}
			if _, err := r.etcdCli.Revoke(context.Background(), r.leaseId); err != nil {
				logs.Error("revoke  lease err:%v", err)
			}
			if r.etcdCli != nil {
				r.etcdCli.Close()
			}
			logs.Info("unregister etcd...")
		case res := <-r.keepAliveCh:
			//	if res != nil {
			//		if err := r.register(); err != nil {
			//			logs.Error(" keepAliveCh register err:%v", err)
			//		}
			//	}
			if res == nil {
				if err := r.register(); err != nil {
					logs.Error("keepAlive register err:%v", err)
				}
				logs.Info("续约成功,%v", res)
			}
		case <-ticker.C:
			if r.keepAliveCh == nil {
				if err := r.register(); err != nil {
					logs.Error("ticker register err:%v", err)
				}
			}
		}
	}

}
func (r *Register) unregister() error {
	_, err := r.etcdCli.Delete(context.Background(), r.info.BuildRegisterKey())
	return err
}
