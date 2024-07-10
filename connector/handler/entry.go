package handler

import (
	"common"
	"common/biz"
	"common/config"
	"common/jwts"
	"common/logs"
	"connector/models/request"
	"context"
	"core/repo"
	"core/service"
	"encoding/json"
	"framework/game"
	"framework/nets"
)

type EntryHandler struct {
	userService *service.UserService
}

func NewEntryHandler(r *repo.Manager) *EntryHandler {
	return &EntryHandler{
		userService: service.NewUserService(r),
	}
}

func (h *EntryHandler) Entry(session *nets.Session, body []byte) (any, error) {
	logs.Info("++++++++++++++++++Entry start+++++++++++++++++++")
	logs.Info("entry request params:%v", string(body))
	logs.Info("++++++++++++++++++Entry end+++++++++++++++++++")
	var req request.EntryReq
	err := json.Unmarshal(body, &req)
	if err != nil {
		return common.Failed(biz.RequestDataError), nil
	}

	//校验token
	uid, err := jwts.ParseToken(req.Token, config.Conf.Jwt.Secret)
	if err != nil {
		logs.Error("parse token err :%v", err)
		return common.Failed(biz.TokenInfoError), nil
	}
	//根据uid，取mongo查询用户，没有则创建、
	user, err := h.userService.FindByUserByUid(context.TODO(), uid, req.UserInfo)
	if err != nil {
		return common.Failed(biz.SqlError), nil
	}
	session.Uid = uid
	return common.Successed(map[string]any{
		"userInfo": user,
		"config":   game.Conf.GetFrontGameConfig(),
	}), nil

}
