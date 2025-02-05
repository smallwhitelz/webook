package repository

import "webook/wire/repository/dao"

type UserRepository struct {
	dao *dao.UserDAO
}

func NewUserRepository(d *dao.UserDAO) *UserRepository {
	return &UserRepository{dao: d}
}
