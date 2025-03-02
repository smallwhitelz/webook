package dao

import (
	"gorm.io/gorm"
	"time"
	"webook/account/domain"
)

func InitTables(db *gorm.DB) error {
	err := db.AutoMigrate(&Account{}, &AccountActivity{})
	if err != nil {
		return err
	}
	// 为了测试和调试方便，这里我补充一个初始化系统账号的代码
	// 你在现实中是不需要的
	now := time.Now().UnixMilli()
	// 忽略这个错误，因为我在测试的反复运行了
	_ = db.Create(&Account{
		Type:     domain.AccountTypeSystem,
		Currency: "CNY",
		Ctime:    now,
		Utime:    now,
	}).Error
	return nil
}
