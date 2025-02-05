package wire

import (
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func InitDB() *gorm.DB {
	db, err := gorm.Open(mysql.Open("xxx"))
	if err != nil {
		panic(err)
	}
	return db
}

func InitRedis() redis.Cmdable {
	return redis.NewClient(&redis.Options{
		Addr: "xxx",
	})
}
