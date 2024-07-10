package dao

import (
	"context"
	"core/repo"
)

const Prefix = "jmqp"
const (
	AccountIdRedisKey string = "AccountId"
	AccountIdBegin    int64  = 10000
	UserIdRedisKey    string = "UserId"
	UserIdBegin       int64  = 100000000
)

type RedisDao struct {
	repo *repo.Manager
}

func (d *RedisDao) NextAccountId() (int64, error) {
	return d.incr(Prefix + ":" + AccountIdRedisKey)
}
func (d *RedisDao) NextUserId() (int64, error) {
	return d.incr(Prefix + ":" + UserIdRedisKey)
}
func (d *RedisDao) incr(keyName string) (int64, error) {
	var exist int64 = 0
	var err error
	if d.repo.Redis.ClusterCli != nil {
		//集群模式
		exist, err = d.repo.Redis.ClusterCli.Exists(context.TODO(), keyName).Result()
	} else {
		exist, err = d.repo.Redis.Cli.Exists(context.TODO(), keyName).Result()
	}
	if exist == 0 {
		var value int64 = 0
		if keyName == Prefix+":"+AccountIdRedisKey {
			value = AccountIdBegin
		}
		if keyName == Prefix+":"+UserIdRedisKey {
			value = UserIdBegin
		}
		var err error = nil
		if d.repo.Redis.ClusterCli != nil {
			err = d.repo.Redis.ClusterCli.Set(context.TODO(), keyName, value, 0).Err()
		} else {
			err = d.repo.Redis.Cli.Set(context.TODO(), keyName, value, 0).Err()
		}
		if err != nil {
			return 0, err
		}
	}
	var id int64 = 0
	if d.repo.Redis.ClusterCli != nil {
		id, err = d.repo.Redis.ClusterCli.Incr(context.TODO(), keyName).Result()
	} else {
		id, err = d.repo.Redis.Cli.Incr(context.TODO(), keyName).Result()
	}
	if err != nil {
		return 0, err
	}
	return id, nil
}

func NewRedisDao(manager *repo.Manager) *RedisDao {
	return &RedisDao{
		repo: manager,
	}
}
