package web

import (
	regexp "github.com/dlclark/regexp2"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
	"net/http"
	"time"
	"webook/internal/domain"
	"webook/internal/errs"
	"webook/internal/service"
	ijwt "webook/internal/web/jwt"
	"webook/pkg/ginx"
)

const (
	emailRegexPattern = "^\\w+([-+.]\\w+)*@\\w+([-.]\\w+)*\\.\\w+([-.]\\w+)*$"
	// 和上面比起来，用 ` 看起来就比较清爽
	passwordRegexPattern = `^(?=.*[A-Za-z])(?=.*\d)(?=.*[$@$!%*#?&])[A-Za-z\d$@$!%*#?&]{8,}$`
	bizLogin             = "login"
)

type UserHandler struct {
	ijwt.Handler
	emailRexExp    *regexp.Regexp
	passwordRexExp *regexp.Regexp
	svc            service.UserService
	codeSvc        service.CodeService
}

func NewUserHandler(svc service.UserService, hdl ijwt.Handler, codeSvc service.CodeService) *UserHandler {
	return &UserHandler{
		emailRexExp:    regexp.MustCompile(emailRegexPattern, regexp.None),
		passwordRexExp: regexp.MustCompile(passwordRegexPattern, regexp.None),
		svc:            svc,
		codeSvc:        codeSvc,
		Handler:        hdl,
	}
}

func (h *UserHandler) RegisterRoutes(server *gin.Engine) {
	// REST 风格
	//server.POST("/user", h.SignUp)
	//server.POST("/user", h.Login)
	//server.GET("/users/:username", h.Profile)
	// 路由分组
	ug := server.Group("/users")
	ug.POST("/signup", ginx.WrapBody(h.SignUp))
	//ug.POST("/login", h.Login)
	ug.POST("/login", ginx.WrapBody(h.LoginJWT))
	ug.POST("/logout", h.LogoutJWT)
	ug.POST("/edit", ginx.WrapBodyAndClaims[EditReq, ijwt.UserClaims](h.Edit))
	ug.GET("/profile", ginx.WrapClaims(h.Profile))
	ug.GET("/refresh_token", h.RefreshToken)
	//ug.GET("/profileSess", h.ProfileSess)

	// 手机验证码登录相关功能
	ug.POST("/login_sms/code/send", h.SendSMSLoginCode)
	ug.POST("/login_sms", ginx.WrapBody(h.LoginSMS))
}

func (h *UserHandler) LoginSMS(ctx *gin.Context, req LoginSMSReq) (ginx.Result, error) {
	ok, err := h.codeSvc.Verify(ctx, bizLogin, req.Phone, req.Code)
	if err != nil {
		zap.L().Error("手机验证码验证失败", zap.Error(err))
		return ginx.Result{
			Code: 5,
			Msg:  "系统异常",
		}, err
	}
	if !ok {
		return ginx.Result{
			Code: 4,
			Msg:  "验证码错误",
		}, nil
	}
	u, err := h.svc.FindOrCreate(ctx, req.Phone)
	if err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "系统异常",
		}, err
	}
	err = h.SetLoginToken(ctx, u.Id)
	if err != nil {
		return ginx.Result{
			Msg:  "系统错误",
			Code: 5,
		}, err
	}
	return ginx.Result{
		Msg: "登录成功",
	}, nil
}

func (h *UserHandler) SendSMSLoginCode(ctx *gin.Context) {
	type Req struct {
		Phone string `json:"phone"`
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		return
	}
	// 这边可以校验req
	if req.Phone == "" {
		ctx.JSON(http.StatusOK, ginx.Result{
			Code: 4,
			Msg:  "请输入手机号码",
		})
		return
	}
	err := h.codeSvc.Send(ctx, bizLogin, req.Phone)
	switch err {
	case nil:
		ctx.JSON(http.StatusOK, ginx.Result{
			Msg: "发送成功",
		})
	case service.ErrCodeSendTooMany:
		// 事实上，防不住有人不知道怎么触发了
		// 少数这种错误是可以接受的
		// 但是频繁出现，那就是有人在搞你的系统
		zap.L().Warn("频繁发送验证码")
		ctx.JSON(http.StatusOK, ginx.Result{
			Code: 4,
			Msg:  "短信发送太频繁，请稍后再试",
		})
	default:
		ctx.JSON(http.StatusOK, ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		})
	}

}

func (h *UserHandler) SignUp(ctx *gin.Context, req SignUpReq) (ginx.Result, error) {
	// 校验邮箱格式
	isEmail, err := h.emailRexExp.MatchString(req.Email)
	if err != nil {
		return ginx.Result{
			Code: errs.UserInternalServerError,
			Msg:  "系统错误",
		}, err
	}
	if !isEmail {
		return ginx.Result{
			Code: errs.UserInvalidInput,
			Msg:  "非法邮箱格式",
		}, nil
	}
	// 两次密码是否一致
	if req.Password != req.ConfirmPassword {
		return ginx.Result{
			Code: errs.UserInvalidInput,
			Msg:  "两次输入的密码不相等",
		}, nil
	}
	// 校验密码
	isPassword, err := h.passwordRexExp.MatchString(req.Password)
	if err != nil {
		return ginx.Result{
			Code: errs.UserInternalServerError,
			Msg:  "系统错误",
		}, err
	}
	if !isPassword {
		return ginx.Result{
			Code: errs.UserInvalidInput,
			Msg:  "密码必须包含字母、数字、特殊字符",
		}, nil
	}

	err = h.svc.Signup(ctx.Request.Context(), domain.User{
		Email:    req.Email,
		Password: req.Password,
	})
	// 判定邮箱冲突
	switch err {
	case nil:
		return ginx.Result{
			Msg: "注册成功",
		}, nil
	case service.ErrDuplicateEmail:
		return ginx.Result{
			Code: errs.UserDuplicateEmail,
			Msg:  "邮箱冲突",
		}, nil
	default:
		return ginx.Result{
			Code: errs.UserInternalServerError,
			Msg:  "系统错误",
		}, err
	}
}

func (h *UserHandler) LoginJWT(ctx *gin.Context, req LoginJWTReq) (ginx.Result, error) {
	u, err := h.svc.Login(ctx, req.Email, req.Password)
	switch err {
	case nil:
		err = h.SetLoginToken(ctx, u.Id)
		if err != nil {
			return ginx.Result{
				Code: errs.UserSetTokenInternalServerError,
				Msg:  "系统错误",
			}, err
		}
		return ginx.Result{
			Msg: "登陆成功",
		}, nil
	case service.ErrInvalidUserOrPassword:
		return ginx.Result{
			Code: errs.UserInvalidOrPassword,
			Msg:  "用户名或密码错误",
		}, err
	default:
		return ginx.Result{
			Code: errs.UserInternalServerError,
			Msg:  "系统错误",
		}, err
	}
}

// 使用session的退出
//func (h *UserHandler) Logout(ctx *gin.Context) {
//	sess := sessions.Default(ctx)
//	sess.Options(sessions.Options{
//		MaxAge: -1,
//	})
//	sess.Save()
//}

func (h *UserHandler) Login(ctx *gin.Context) {
	type Req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	var req Req
	if err := ctx.Bind(&req); err != nil {
		return
	}
	u, err := h.svc.Login(ctx, req.Email, req.Password)
	switch err {
	case nil:
		sess := sessions.Default(ctx)
		sess.Set("userId", u.Id)
		sess.Options(sessions.Options{
			// 900s = 15m
			MaxAge: 900,
		})
		err := sess.Save()
		if err != nil {
			ctx.String(http.StatusOK, "系统错误")
			return
		}
		ctx.String(http.StatusOK, "登录成功")
	case service.ErrInvalidUserOrPassword:
		ctx.String(http.StatusOK, "用户名或密码错误")
	default:
		ctx.String(http.StatusOK, "系统错误")
	}
}

func (h *UserHandler) Edit(ctx *gin.Context, req EditReq, uc ijwt.UserClaims) (ginx.Result, error) {
	// 用户生日格式输入不对
	birthday, err := time.Parse(time.DateOnly, req.Birthday)
	if err != nil {
		return ginx.Result{
			Code: errs.UserInvalidInput,
			Msg:  "生日格式不正确",
		}, err
	}
	err = h.svc.UpdateNonSensitiveInfo(ctx, domain.User{
		Id:          uc.Uid,
		Nickname:    req.Nickname,
		Birthday:    birthday,
		Description: req.Description,
	})
	if err != nil {
		return ginx.Result{
			Msg: "修改失败",
		}, err
	}
	return ginx.Result{
		Msg: "更新成功",
	}, nil
}

func (h *UserHandler) Profile(ctx *gin.Context, uc ijwt.UserClaims) (ginx.Result, error) {
	u, err := h.svc.FindById(ctx, uc.Uid)
	if err != nil {
		return ginx.Result{
			Msg: "系统错误",
		}, err
	}
	type User struct {
		Nickname    string `json:"nickname"`
		Email       string `json:"email"`
		Birthday    string `json:"birthday"`
		Description string `json:"description"`
	}
	return ginx.Result{
		Msg: "获取成功",
		Data: User{
			Nickname:    u.Nickname,
			Email:       u.Email,
			Birthday:    u.Birthday.Format(time.DateOnly),
			Description: u.Description,
		},
	}, nil
}

// ProfileSess 使用 session 机制的 Profile
func (h *UserHandler) ProfileSess(ctx *gin.Context) {
	sess := sessions.Default(ctx)
	idVal := sess.Get("userId")
	uid, ok := idVal.(int64)
	//绝大部分的情况下是因为代码出了问题
	if !ok {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	u, err := h.svc.FindById(ctx, uid)
	if err != nil {
		ctx.String(http.StatusOK, "系统错误")
		return
	}
	type User struct {
		Nickname    string `json:"nickname"`
		Email       string `json:"email"`
		Birthday    string `json:"birthday"`
		Description string `json:"description"`
	}
	ctx.JSON(http.StatusOK, User{
		Nickname:    u.Nickname,
		Email:       u.Email,
		Birthday:    u.Birthday.Format(time.DateOnly),
		Description: u.Description,
	})
}

func (h *UserHandler) RefreshToken(ctx *gin.Context) {
	// 约定 前端在Authorization里面带上 refresh_token
	tokenStr := h.ExtractToken(ctx)
	var rc ijwt.RefreshClaims
	token, err := jwt.ParseWithClaims(tokenStr, &rc, func(token *jwt.Token) (interface{}, error) {
		return ijwt.RCJWTKey, nil
	})
	if err != nil {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	if token == nil || !token.Valid {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	err = h.CheckSession(ctx, rc.Ssid)
	if err != nil {
		// token无效或者 redis有问题
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	err = h.SetJWTToken(ctx, rc.Uid, rc.Ssid)
	if err != nil {
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	ctx.JSON(http.StatusOK, ginx.Result{Msg: "OK"})
}

func (h *UserHandler) LogoutJWT(ctx *gin.Context) {
	err := h.ClearToken(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		})
	}
	ctx.JSON(http.StatusOK, ginx.Result{
		Msg: "退出登录成功",
	})

}
