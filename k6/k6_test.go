package k6

import (
	"github.com/gin-gonic/gin"
	"math/rand"
	"net/http"
	"testing"
	"time"
)

func TestHello(t *testing.T) {
	server := gin.Default()
	server.POST("/hello", func(ctx *gin.Context) {
		var u User
		if err := ctx.Bind(&u); err != nil {
			return
		}
		// 在这边模拟一下业务
		// [0,1000)
		number := rand.Int31n(1000)
		// 模拟响应时间，睡眠
		time.Sleep(time.Millisecond * time.Duration(number))
		// 模拟一个 10% 的错误比率
		if number%100 < 10 {
			ctx.String(http.StatusInternalServerError, "模拟服务器失败")
		} else {
			ctx.String(http.StatusOK, u.Name)
		}
	})
	server.Run(":8080")
}

type User struct {
	Name string `json:"name"`
}
