package dao

import (
	"context"
	"errors"
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

// Preempt 考虑到了续约失败的场景
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
		// 先找到可以被强占的job
		err := db.Where("(status = ? AND next_time <?) OR (status = ? AND utime < ?)",
			jobStatusWaiting, now, jobStatusRunning, ddl).
			First(&j).Error
		if err != nil {
			return j, err
		}

		//
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

// PreemptV2 没有考虑续约失败的情景
func (dao *GORMJobDAO) PreemptV2(ctx context.Context) (Job, error) {
	db := dao.db.WithContext(ctx)
	for {
		var j Job
		now := time.Now().UnixMilli()
		// next_time < ? 所有处于waiting状态的不一定立刻要执行，比如我有些job是3分钟后，所以这里也要进行判断
		// 这里可以解释为nextTime是这个job下一次可以执行的时间，
		// 如果nexttime<当前时间，说明当前时间已经晚于job下一次执行的时间，那么就可以抢来用
		err := db.Where("status = ? AND next_time < ?", jobStatusWaiting, now).First(&j).Error
		if err != nil {
			return j, err
		}

		res := db.Model(&Job{}).Where("id = ? AND version = ?", j.Id, j.Version).Updates(map[string]any{
			"status":  jobStatusRunning,
			"version": j.Version + 1,
			"utime":   now,
		})
		if res.Error != nil {
			// 抢占本身出了问题，那就应该返回，因为你本身也做不了什么
			return Job{}, res.Error
		}
		if res.RowsAffected == 0 {
			// 没抢到
			// 这里没抢到不应该直接返回，我抢不到一条我可以继续抢
			//return Job{}, errors.New("没抢到")
			continue
		}
		return j, err
	}
}

// PreemptV1 这种写法会出现并发安全问题，可能存在多个goroutine去抢
func (dao *GORMJobDAO) PreemptV1(ctx context.Context) (Job, error) {
	var j Job
	err := dao.db.WithContext(ctx).Where("status = ?", jobStatusWaiting).First(&j).Error
	if err != nil {
		return j, err
	}
	res := dao.db.WithContext(ctx).Model(&Job{}).Where("id = ?", j.Id).Updates(map[string]any{
		"status": jobStatusRunning,
	})
	if res.RowsAffected == 0 {
		// 没抢到
		return Job{}, errors.New("没抢到")
	}
	return j, err
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

	// 乐观锁机制，版本
	Version int

	// 单独建立索引，因为NextTime过滤后搜出来的数据就很少，并且搜出来的大多是waiting状态
	// 如果这个时候再给status建立联合索引有些多此一举
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
