package service

import (
	"common/biz"
	"context"
	"core/dao"
	"core/models/entity"
	"core/models/requests"
	"core/repo"
	"framework/msError"
	"strconv"
	"time"
	"user/pb"
)

//创建账号

type AccountService struct {
	redisDao   *dao.RedisDao
	accountDao *dao.AccountDao
	pb.UnimplementedUserServiceServer
}

func NewAccountService(manager *repo.Manager) *AccountService {
	return &AccountService{
		accountDao: dao.NewAccountDao(manager),
		redisDao:   dao.NewRedisDao(manager),
	}
}

func (a *AccountService) Register(ctx context.Context, req *pb.RegisterParams) (*pb.RegisterResponse, error) {
	//注册逻辑
	if req.LoginPlatform == requests.WeiXin {
		ac, err := a.wxRegister(req)
		if err != nil {
			return &pb.RegisterResponse{}, msError.GrpcError(err)
		}
		return &pb.RegisterResponse{
			Uid: ac.Uid,
		}, nil
	}

	return &pb.RegisterResponse{}, nil
}
func (a *AccountService) wxRegister(req *pb.RegisterParams) (*entity.Account, *msError.Error) {
	//封装account结构，存入数据库 Mongo 分布式id，objectID
	ac := &entity.Account{
		WxAccount:  req.Account,
		CreateTime: time.Now(),
	}
	uid, err := a.redisDao.NextAccountId()
	if err != nil {
		return ac, biz.SqlError
	}
	ac.Uid = strconv.FormatInt(uid, 10)
	err = a.accountDao.SaveAccount(context.TODO(), ac)
	if err != nil {
		return ac, biz.SqlError

	}
	return ac, nil
}
