//go:build wireinject

package wire

import (
	"github.com/google/wire"
	"webook/wire/repository"
	"webook/wire/repository/dao"
)

func InitUserRepository() *repository.UserRepository {
	wire.Build(repository.NewUserRepository, dao.NewUserDAO, InitDB, InitRedis)
	return &repository.UserRepository{}
}
