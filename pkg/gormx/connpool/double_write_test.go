package connpool

import (
	"github.com/ecodeclub/ekit/syncx/atomicx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"testing"
	"webook/pkg/logger"
)

type DoubleWriteTestSuite struct {
	suite.Suite
	db *gorm.DB
}

func (s *DoubleWriteTestSuite) SetupSuite() {
	t := s.T()
	src, err := gorm.Open(mysql.Open("root:root@tcp(43.154.97.245:13316)/webook"))
	require.NoError(t, err)
	err = src.AutoMigrate(&Interactive{})
	require.NoError(t, err)
	dst, err := gorm.Open(mysql.Open("root:root@tcp(43.154.97.245:13316)/webook_intr"))
	require.NoError(t, err)
	err = dst.AutoMigrate(&Interactive{})
	require.NoError(t, err)

	doubleWrite, err := gorm.Open(mysql.New(mysql.Config{
		Conn: &DoubleWritePool{
			src:     src.ConnPool,
			dst:     dst.ConnPool,
			pattern: atomicx.NewValueOf(PatternSrcFirst),
			l:       logger.NewNoOpLogger(),
		},
	}))
	require.NoError(t, err)
	s.db = doubleWrite
}

func (s *DoubleWriteTestSuite) TearDownTest() {
	//s.db.Exec("TRUNCATE TABLE interactives")
}

// 集成测试，需要启动数据库
func (s *DoubleWriteTestSuite) TestDoubleWriteTest() {
	t := s.T()
	err := s.db.Create(&Interactive{
		Biz:   "test",
		BizId: 10086,
	}).Error
	assert.NoError(t, err)
	// 查询数据库就可以看到对应的数据
}

func (s *DoubleWriteTestSuite) TestDoubleWriteTransaction() {
	t := s.T()
	err := s.db.Transaction(func(tx *gorm.DB) error {
		return tx.Create(&Interactive{
			Biz:   "test",
			BizId: 10087,
		}).Error
	})
	require.NoError(t, err)
}

func TestDoubleWrite(t *testing.T) {
	suite.Run(t, new(DoubleWriteTestSuite))
}

type Interactive struct {
	Id         int64  `gorm:"primaryKey,autoIncrement"`
	BizId      int64  `gorm:"uniqueIndex:biz_type_id"`
	Biz        string `gorm:"type:varchar(128);uniqueIndex:biz_type_id"`
	ReadCnt    int64
	CollectCnt int64
	LikeCnt    int64
	Ctime      int64
	Utime      int64
}
