package service

import (
	"common/biz"
	"common/logs"
	"common/utils"
	"connector/models/request"
	"context"
	"core/dao"
	"core/models/entity"
	"core/repo"
	"fmt"
	"framework/game"
	"framework/msError"
	hall "hall/models/request"
	"time"
)

type UserService struct {
	userDao *dao.UserDao
}

func NewUserService(r *repo.Manager) *UserService {
	return &UserService{
		userDao: dao.NewUserDao(r),
	}
}

func (s *UserService) FindByUserByUid(ctx context.Context, uid string, info request.UserInfo) (*entity.User, error) {
	//查询mongo
	user, err := s.userDao.FindByUserByUid(ctx, uid)
	if err != nil {
		logs.Error("FindByUserByUid err: %v", err)
		return nil, err
	}

	if user == nil {
		//新增
		user := &entity.User{}
		user.Uid = uid
		user.Gold = int64(game.Conf.GameConfig["startGold"]["value"].(float64))
		user.Avatar = utils.Default(info.Avatar, "Common/head_icon_default")
		user.Nickname = utils.Default(info.Nickname, fmt.Sprintf("%s%s", "赌神", uid))
		user.Sex = info.Sex //0男1女
		user.CreateTime = time.Now().UnixMilli()
		user.LastLoginTime = time.Now().UnixMilli()
		err := s.userDao.Insert(context.TODO(), user)
		if err != nil {
			logs.Error("Insert user err:%v", err)
			return nil, err
		}
	}
	return user, nil
}

func (s *UserService) UpdateUserAddressByUid(uid string, req hall.UpdateUserAddressReq) error {
	user := &entity.User{
		Uid:      uid,
		Address:  req.Address,
		Location: req.Location,
	}
	err := s.userDao.UpdateUserAddressByUid(context.TODO(), user)
	if err != nil {
		logs.Error("UpdateUserAddressByUid err: %v", err)
		return err
	}
	return nil
}

func (s *UserService) FindUserByUid(ctx context.Context, uid string) (*entity.User, *msError.Error) {
	//查询mongo
	user, err := s.userDao.FindByUserByUid(ctx, uid)
	if err != nil {
		logs.Error("FindUserByUid err: %v", err)
		return nil, biz.SqlError
	}

	return user, nil

}
