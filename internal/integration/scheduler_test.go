package integration

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
	"testing"
	"time"
	"webook/internal/domain"
	"webook/internal/integration/startup"
	"webook/internal/job"
	"webook/internal/repository/dao"
)

type SchedulerTestSuite struct {
	suite.Suite
	scheduler *job.Scheduler
	db        *gorm.DB
}

func (s *SchedulerTestSuite) SetupSuite() {
	s.db = startup.InitDB()
	s.scheduler = startup.InitJobScheduler()
}

func (s *SchedulerTestSuite) TearDownSuite() {
	err := s.db.Exec("TRUNCATE TABLE `jobs`").Error
	assert.NoError(s.T(), err)
}

// TestSchedule 测试调度
func (s *SchedulerTestSuite) TestSchedule() {
	testCases := []struct {
		name   string
		before func(t *testing.T)
		after  func(t *testing.T)

		// Start 运行时间
		interval time.Duration
		wantErr  error
		wantJob  *testJob
	}{
		{
			name: "验证调度",
			before: func(t *testing.T) {
				// 在db里插入一条job数据
				j := dao.Job{
					Id:       1,
					Name:     "test_job",
					Executor: "local",
					Cfg:      "这是一个配置",
					// 每五秒执行一次
					Expression: "*/5 * * * * ?",
					NextTime:   time.Now().UnixMilli(),
					Ctime:      123,
					Utime:      234,
				}
				err := s.db.Create(&j).Error
				assert.NoError(t, err)
			},
			after: func(t *testing.T) {
				var j dao.Job
				err := s.db.Where("id = ?", 1).First(&j).Error
				assert.NoError(t, err)
				// 应该大于当下
				assert.True(t, j.NextTime > time.Now().UnixMilli())
				j.NextTime = 0
				assert.True(t, j.Ctime > 0)
				j.Ctime = 0
				assert.True(t, j.Utime > 0)
				j.Utime = 0
				assert.Equal(t, dao.Job{
					Id:         1,
					Name:       "test_job",
					Executor:   "local",
					Expression: "*/5 * * * * ?",
					Cfg:        "这是一个配置",
					Status:     0,
					// 抢占导致版本升高
					Version: 1,
				}, j)
			},
			wantErr: context.DeadlineExceeded,
			// 运行了一次
			wantJob: &testJob{cnt: 1},
			// 开始调度一秒钟
			interval: time.Second,
		},
	}
	for _, tc := range testCases {
		s.T().Run(tc.name, func(t *testing.T) {
			tc.before(t)
			defer tc.after(t)
			exec := job.NewLocalFuncExecutor()
			j := &testJob{}
			exec.RegisterFunc("test_job", j.Do)
			s.scheduler.RegisterExecutor(exec)
			ctx, cancel := context.WithTimeout(context.Background(), tc.interval)
			defer cancel()
			err := s.scheduler.Schedule(ctx)
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.wantJob, j)
		})
	}
}

func TestScheduler(t *testing.T) {
	suite.Run(t, &SchedulerTestSuite{})
}

type testJob struct {
	cnt int
}

func (t *testJob) Do(ctx context.Context, j domain.Job) error {
	t.cnt++
	return nil
}
