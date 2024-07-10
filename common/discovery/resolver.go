package discovery

import (
	"common/config"
	"common/logs"
	"context"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc/attributes"
	"google.golang.org/grpc/resolver"
	"time"
)

type Resolver struct {
	DialTimeout int              //超时时间
	etcdCli     *clientv3.Client //etcd连接
	closeCh     chan struct{}
	conf        config.EtcdConf
	key         string
	cc          resolver.ClientConn
	srvAddrList []resolver.Address
	watchCh     clientv3.WatchChan
}

func (r Resolver) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	r.cc = cc
	//连接etcd
	var err error
	r.etcdCli, err = clientv3.New(clientv3.Config{
		Endpoints:   r.conf.Addrs,
		DialTimeout: time.Duration(r.DialTimeout) * time.Second,
	})
	if err != nil {
		logs.Fatal("create etcd resolver client error:", err)
	}
	r.closeCh = make(chan struct{})
	//根据key获取value
	r.key = target.URL.Path
	if err := r.sync(); err != nil {
		return nil, err
	}
	go r.watch()
	return nil, nil

}

func (r Resolver) Scheme() string {
	return "etcd"
}

func (r Resolver) sync() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(r.conf.RWTimeout)*time.Second)
	defer cancel()
	res, err := r.etcdCli.Get(ctx, r.key, clientv3.WithPrefix())
	if err != nil {
		logs.Error("get etcd key error:", r.key, err)
		return err
	}
	//logs.Info("%v", res.Kvs)
	r.srvAddrList = []resolver.Address{}
	for _, v := range res.Kvs {
		server, err := ParseValue(v.Value)
		if err != nil {
			logs.Error("parse value error:", v.Key, err)
			continue
		}
		r.srvAddrList = append(r.srvAddrList, resolver.Address{
			Addr:       server.Addr,
			Attributes: attributes.New("weight", server.Weight),
		})
	}
	if len(r.srvAddrList) == 0 {
		logs.Error("get etcd addr error:", r.key)
		return nil
	}
	err = r.cc.UpdateState(resolver.State{
		Addresses: r.srvAddrList,
	})
	if err != nil {
		logs.Error("update etcd resolver state error:", r.key, err)
	}
	return nil
}
func NewResolver(conf config.EtcdConf) *Resolver {
	return &Resolver{
		conf:        conf,
		DialTimeout: conf.DialTimeout,
	}
}
func (r Resolver) watch() {
	//定时同步数据
	//监听事件
	//监听close，关闭etcd
	r.etcdCli.Watch(context.Background(), r.key, clientv3.WithPrefix(), clientv3.WithPrevKV())
	ticker := time.NewTicker(time.Minute)
	for {
		select {
		case <-r.closeCh:
			r.Close()
		case res, ok := <-r.watchCh:
			if ok {
				//
				r.update(res.Events)

			}
		case <-ticker.C:
			if err := r.sync(); err != nil {
				logs.Error("update etcd resolver state error:", r.key, err)
			}
		}
	}

}
func (r Resolver) update(events []*clientv3.Event) {
	for _, event := range events {
		switch event.Type {
		case clientv3.EventTypePut:
			server, err := ParseValue(event.Kv.Value)
			if err != nil {
				logs.Error("parse value error:", event.Kv.Key, err)
			}
			address := resolver.Address{
				Addr:       server.Addr,
				Attributes: attributes.New("weight", server.Weight),
			}
			if !Exist(r.srvAddrList, address) {
				r.srvAddrList = append(r.srvAddrList, address)
				err = r.cc.UpdateState(resolver.State{
					Addresses: r.srvAddrList,
				})
				if err != nil {
					logs.Error("update etcd resolver state error:", r.key, err)
				}
			}
		case clientv3.EventTypeDelete:
			//接收到delete操作，删除r.survivalist其中匹配的
			server, err := ParseKey(string(event.Kv.Key))
			if err != nil {
				logs.Error("parse value error:", event.Kv.Key, err)
			}
			addr := resolver.Address{
				Addr: server.Addr,
			}
			if list, ok := Remove(r.srvAddrList, addr); ok {
				r.srvAddrList = list
				err := r.cc.UpdateState(resolver.State{
					Addresses: r.srvAddrList,
				})
				if err != nil {
					logs.Error("update etcd resolver state error:", r.key, err)
				}
			}
		}
	}
}
func Remove(list []resolver.Address, addr resolver.Address) ([]resolver.Address, bool) {
	for i := range list {
		if list[i].Addr == addr.Addr {
			list[i] = list[len(list)-1]
			return list[:len(list)-1], true
		}
	}
	return nil, false
}
func Exist(list []resolver.Address, addr resolver.Address) bool {
	for i := range list {
		if list[i].Addr == addr.Addr {
			return true
		}
	}
	return false
}
func (r Resolver) Close() {
	if r.etcdCli != nil {
		err := r.etcdCli.Close()
		if err != nil {
			logs.Error("close etcd client error:", err)
		}
	}

}
