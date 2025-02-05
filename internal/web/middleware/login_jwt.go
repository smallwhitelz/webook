package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"net/http"
	ijwt "webook/internal/web/jwt"
)

type LoginJWTMiddlewareBuilder struct {
	ijwt.Handler
}

func NewLoginJWTMiddlewareBuilder(hdl ijwt.Handler) *LoginJWTMiddlewareBuilder {
	return &LoginJWTMiddlewareBuilder{
		Handler: hdl,
	}
}

func (m *LoginJWTMiddlewareBuilder) CheckLogin() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		path := ctx.Request.URL.Path
		if path == "/users/signup" ||
			path == "/users/login" ||
			path == "/users/refresh_token" ||
			path == "/users/login_sms/code/send" ||
			path == "/users/login_sms" ||
			path == "/oauth2/wechat/authurl" ||
			path == "/oauth2/wechat/callback" {
			// 直接放行
			return
		}

		tokenStr := m.ExtractToken(ctx)
		var uc ijwt.UserClaims
		token, err := jwt.ParseWithClaims(tokenStr, &uc, func(token *jwt.Token) (interface{}, error) {
			return ijwt.JWTKey, nil
		})
		if err != nil {
			// token不对，token是伪造的
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		if token == nil || !token.Valid {
			// 在这里发现access_token过期了，生成一个新的 access_token

			// token解析出来了，但是token是非法的，或者过期了
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		if uc.UserAgent != ctx.GetHeader("User-Agent") {
			// 后期接入监控告警，这个地方要埋点，也就是记录下来
			//能够进这个分支的，说明大概率是攻击者
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		//expireTime := uc.ExpiresAt
		//if expireTime.Before(time.Now()) {
		//	ctx.AbortWithStatus(http.StatusUnauthorized)
		//	return
		//}
		// 剩余过期时间<50s 就要刷新
		//if expireTime.Sub(time.Now()) < time.Second*50 {
		//	uc.ExpiresAt = jwt.NewNumericDate(time.Now().Add(time.Minute * 30))
		//	tokenStr, err = token.SignedString(web.JWTKey)
		//	ctx.Header("x-jwt-token", tokenStr)
		//	if err != nil {
		//		// 这边仅仅是过期时间没有刷新，但是用户是登录了的
		//		log.Println(err)
		//	}
		//}

		// 这里看
		err = m.CheckSession(ctx, uc.Ssid)
		if err != nil {
			// token无效或者 redis有问题
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// 可以兼容 Redis异常情况
		// 做好监控，监控有没有error
		//if cnt > 0 {
		//	// token无效或者 redis有问题
		//	ctx.AbortWithStatus(http.StatusUnauthorized)
		//	return
		//}
		ctx.Set("user", uc)
	}
}
