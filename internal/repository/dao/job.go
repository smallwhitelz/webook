package dao

import (
	"context"
	"gorm.io/gorm"
	"time"
)

type JobDAO interface {
	Preempt(ctx context.Context) (Job, error)
	Release(ctx context.Context, jid int64) error
	UpdateUtime(ctx context.Context, id int64) error
	UpdateNextTime(ctx context.Context, jid int64, t time.Time) error
}

type GORMJobDAO struct {
	db *gorm.DB
}

func NewGORMJobDAO(db *gorm.DB) JobDAO {
	return &GORMJobDAO{db: db}
}

func (dao *GORMJobDAO) Preempt(ctx context.Context) (Job, error) {
	db := dao.db.WithContext(ctx)
	for {
		var j Job
		now := time.Now().UnixMilli()
		// 假设说续约是

		// 作业：这里是缺少找到续约失败的 JOB 出来执行
		// 续约失败就是本来应该续约
		// 而如果续约成功了，那么肯定 utime 必然在一个范围内
		// 比如说，续约是一分钟，那么 utime 距离当下，必然在一分钟内
		// 我们可以说连续 utime < 当前三分钟前，就认为续约失败了
		ddl := now - (time.Minute * 3).Milliseconds()
		err := db.Where("(status = ? AND next_time <?) OR (status = ? AND utime < ?)",
			jobStatusWaiting, now, jobStatusRunning, ddl).
			First(&j).Error
		if err != nil {
			return j, err
		}

		res := db.Model(&Job{}).
			Where("id = ? AND version = ?", j.Id, j.Version).Updates(map[string]any{
			"status":  jobStatusRunning,
			"version": j.Version + 1,
			"utime":   now,
		})
		if res.Error != nil {
			return Job{}, res.Error
		}
		if res.RowsAffected == 0 {
			// 没抢到
			continue
		}
		return j, err
	}
}

func (dao *GORMJobDAO) Release(ctx context.Context, jid int64) error {
	now := time.Now().UnixMilli()
	return dao.db.WithContext(ctx).Model(&Job{}).Where("id = ?", jid).Updates(map[string]any{
		"status": jobStatusWaiting,
		"utime":  now,
	}).Error
}

func (dao *GORMJobDAO) UpdateUtime(ctx context.Context, jid int64) error {
	now := time.Now().UnixMilli()
	return dao.db.WithContext(ctx).Model(&Job{}).Where("id = ?", jid).Updates(map[string]any{
		"utime": now,
	}).Error
}

func (dao *GORMJobDAO) UpdateNextTime(ctx context.Context, jid int64, t time.Time) error {
	now := time.Now().UnixMilli()
	return dao.db.WithContext(ctx).Model(&Job{}).Where("id = ?", jid).Updates(map[string]any{
		"utime":     now,
		"next_time": t.UnixMilli(),
	}).Error
}

type Job struct {
	Id         int64  `gorm:"primaryKey,autoIncrement"`
	Name       string `gorm:"type:varchar(128);unique"`
	Executor   string
	Expression string
	Cfg        string

	// 状态来表达，是不是可以抢占，有没有被人抢占
	Status int

	Version int

	NextTime int64 `gorm:"index"`

	Ctime int64
	Utime int64
}

const (
	// jobStatusWaiting 没人抢
	jobStatusWaiting = iota
	// jobStatusRunning 已经被人抢了
	jobStatusRunning
	// jobStatusPaused 不再需要调度了
	jobStatusPaused
)
