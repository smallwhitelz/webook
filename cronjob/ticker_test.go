package cronjob

import (
	"context"
	"testing"
	"time"
)

func TestTicker(t *testing.T) {
	// 间隔一秒的ticker
	ticker := time.NewTicker(time.Second)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	// 记住要停掉ticker
	defer ticker.Stop()
	// 每隔一秒钟就会有一个信号
	for {
		select {
		case <-ctx.Done():
			// 循环结束
			t.Log("循环结束")
			// 结束掉也可以return，select中使用break无法中断for循环
			// 尽量不要用goto，可读性很差
			goto end
		case now := <-ticker.C:
			t.Log("过了一秒", now.UnixMilli())

		}
	}
end:
	t.Log("goto 过来了，结束程序")
}
