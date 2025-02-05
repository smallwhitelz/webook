package startup

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	dao2 "webook/interactive/repository/dao"
	"webook/internal/repository/dao"
)

func InitDB() *gorm.DB {
	db, err := gorm.Open(mysql.Open("root:root@tcp(43.154.97.245:13316)/webook"), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	err = dao.InitTables(db)
	if err != nil {
		panic(err)
	}
	err = dao2.InitTables(db)
	if err != nil {
		panic(err)
	}
	return db
}
