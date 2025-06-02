package dao

import "gorm.io/gorm"

func InitTables(db *gorm.DB) error {
	// 严格来说，这不是优秀的实践
	return db.AutoMigrate(
		&Interactive{},
		&UserLikeBiz{},
		&UserCollectionBiz{},
	)
}
