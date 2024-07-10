package entity

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

//用户密码登录，生成账号，账号下可能有多个角色

type Account struct {
	Id           primitive.ObjectID `bson:"_id,omitempty"`
	Uid          string             `bson:"uid"`
	Account      string             `bson:"account" `
	Password     string             `bson:"password"`
	PhoneAccount string             `bson:"phoneAccount"`
	WxAccount    string             `bson:"wxAccount"`
	CreateTime   time.Time          `bson:"createTime"`
}
