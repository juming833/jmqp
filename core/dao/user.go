package dao

import (
	"context"
	"core/models/entity"
	"core/repo"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserDao struct {
	repo *repo.Manager
}

func NewUserDao(m *repo.Manager) *UserDao {
	return &UserDao{
		repo: m,
	}

}

func (d *UserDao) FindByUserByUid(ctx context.Context, uid string) (*entity.User, error) {
	db := d.repo.Mongo.Db.Collection("user")
	singleResult := db.FindOne(ctx, bson.D{
		{"uid", uid},
	})
	user := new(entity.User)
	err := singleResult.Decode(user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return user, nil
}
func (d *UserDao) Insert(ctx context.Context, user *entity.User) error {
	db := d.repo.Mongo.Db.Collection("user")
	_, err := db.InsertOne(ctx, user)
	return err
}

func (d *UserDao) UpdateUserAddressByUid(ctx context.Context, user *entity.User) error {
	db := d.repo.Mongo.Db.Collection("user")
	_, err := db.UpdateOne(ctx, bson.M{
		"uid": user.Uid,
	}, bson.M{
		"$set": bson.M{
			"address": user.Address,
			"Locatin": user.Location,
		},
	})
	if err != nil {
		return err
	}
	return nil
}
