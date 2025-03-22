package repository

import (
	"context"
	"github.com/ecodeclub/ekit/slice"
	"webook/search/domain"
	"webook/search/repository/dao"
)

type userRepository struct {
	dao dao.UserDAO
}

func (u *userRepository) SearchUser(ctx context.Context, keywords []string) ([]domain.User, error) {
	users, err := u.dao.Search(ctx, keywords)
	if err != nil {
		return nil, err
	}
	return slice.Map(users, func(idx int, src dao.User) domain.User {
		return domain.User{
			Id:       src.Id,
			Email:    src.Email,
			Nickname: src.Nickname,
			Phone:    src.Phone,
		}
	}), nil
}

func (u *userRepository) InputUser(ctx context.Context, user domain.User) error {
	return u.dao.InputUser(ctx, dao.User{
		Id:       user.Id,
		Email:    user.Email,
		Nickname: user.Nickname,
		Phone:    user.Phone,
	})
}

func NewUserRepository(dao dao.UserDAO) UserRepository {
	return &userRepository{dao: dao}
}
