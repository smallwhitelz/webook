package middleware

import (
	"encoding/gob"
	"fmt"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

type LoginMiddlewareBuilder struct {
}

func (m *LoginMiddlewareBuilder) CheckLogin() gin.HandlerFunc {
	// 注册一下这个类型,因为time是go里面的结构体，redis是用C语言写的，所以需要注册一下，才能把time序列化
	// 如果不序列化，那么redis里面存储的会是内存地址，那么redis里面取出来，就是0xc00000e0b0，是看不懂的
	// 所以我们需要把time序列化成字符串，这样redis里面存储的就是字符串了，redis可以直接读取，这样的话就可以刷新登陆状态了
	gob.Register(time.Now())
	return func(ctx *gin.Context) {
		path := ctx.Request.URL.Path
		if path == "/users/signup" || path == "/users/login" {
			// 直接放行
			return
		}
		sess := sessions.Default(ctx)
		userId := sess.Get("userId")
		if userId == nil {
			// 中断，不要执行后面的业务逻辑了
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		now := time.Now()
		// 刷新登陆状态： 我们怎么知道要刷新了呢
		// 假如说，我们的策略是每分钟刷新一次，我们怎么知道已经过了一分钟？
		const updateTimeKey = "update_time"
		val := sess.Get(updateTimeKey)
		lastUpdateTime, ok := val.(time.Time) // 类型断言，断言val的类型是time.Time，如果是，那么ok就是true，否则就是false
		// 如果没有获取到val，并且也不是time类型并且超过了一分钟，那么就要重新设置sess
		if val == nil || (!ok) || now.Sub(lastUpdateTime) > time.Minute {
			// 第一登陆
			sess.Set(updateTimeKey, now)
			//sessions是覆盖性的。所以上面的时间设置完，我们要再设置一次userId
			sess.Set("userId", userId)
			err := sess.Save()
			if err != nil {
				// 打日志,不中断，因为不影响业务使用
				fmt.Println(err)
			}
		}
	}
}
