package cronjob

import (
	"github.com/robfig/cron/v3"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestCronExpr(t *testing.T) {
	expr := cron.New(cron.WithSeconds())
	id, err := expr.AddFunc("@every 1s", func() {
		t.Log("执行了")
	})
	assert.NoError(t, err)
	t.Log("任务", id)
	expr.Start()
	time.Sleep(time.Second * 10)
	ctx := expr.Stop() // 意思是你不要调度新任务执行了，你正在执行的继续执行
	t.Log("发出来停止信号")
	<-ctx.Done()
	t.Log("彻底停下来了，没有任务在执行")
	// 这边，彻底停下来
}

type JobFunc func()

func (j JobFunc) Run() {
	j()
}
