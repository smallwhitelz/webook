package web

import (
	"fmt"
	"github.com/ecodeclub/ekit/slice"
	"github.com/gin-gonic/gin"
	"golang.org/x/sync/errgroup"
	"net/http"
	"strconv"
	"time"
	"webook/internal/domain"
	"webook/internal/service"
	"webook/internal/web/jwt"
	"webook/pkg/ginx"
	"webook/pkg/logger"
)

type ArticleHandler struct {
	svc     service.ArticleService
	intrSvc service.InteractiveService
	l       logger.LoggerV1
	biz     string
}

func NewArticleHandler(svc service.ArticleService, l logger.LoggerV1, intrSvc service.InteractiveService) *ArticleHandler {
	return &ArticleHandler{
		svc:     svc,
		l:       l,
		intrSvc: intrSvc,
		biz:     "article",
	}
}

func (h *ArticleHandler) RegisterRoutes(server *gin.Engine) {
	g := server.Group("/articles")
	// 新增和修改接口
	g.POST("/edit", ginx.WrapBodyAndClaims(h.Edit))

	// 发表接口
	g.POST("/publish", ginx.WrapBodyAndClaims(h.Publish))
	g.POST("/withdraw", ginx.WrapBodyAndClaims(h.Withdraw))

	// 创作者接口
	// 返回文章详情
	g.GET("/detail/:id", h.Detail)
	// 按道理来说，这边就是get方法
	// /list?offset=?&limit=?
	// 返回文章列表
	g.POST("/list", h.List)

	// 读者接口
	pub := g.Group("/pub")
	pub.GET("/:id", h.PubDetail)
	// 传入一个参数，true就是点赞，false就是不点赞
	pub.POST("/like", ginx.WrapBodyAndClaims(h.Like))
	pub.POST("/collect", ginx.WrapBodyAndClaims(h.Collect))
}

// Edit 接收Article 输入，返回一个ID，文章ID
func (h *ArticleHandler) Edit(ctx *gin.Context, req ArticleEditReq, uc jwt.UserClaims) (ginx.Result, error) {
	id, err := h.svc.Save(ctx, domain.Article{
		Id:      req.Id,
		Title:   req.Title,
		Content: req.Content,
		Author: domain.Author{
			Id: uc.Uid,
		},
	})
	if err != nil {
		h.l.Error("保存文章数据失败",
			logger.Int64("uid", uc.Uid),
			logger.Error(err))
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, err
	}
	return ginx.Result{
		Data: id,
	}, nil
}

func (h *ArticleHandler) Publish(ctx *gin.Context, req PublishReq, uc jwt.UserClaims) (ginx.Result, error) {
	id, err := h.svc.Publish(ctx, domain.Article{
		Id:      req.Id,
		Title:   req.Title,
		Content: req.Content,
		Author: domain.Author{
			Id: uc.Uid,
		},
	})
	if err != nil {
		return ginx.Result{
			Msg:  "系统错误",
			Code: 5,
		}, fmt.Errorf("发表文章失败 aid %d, uid %d %w", req.Id, uc.Uid, err)
	}
	return ginx.Result{
		Data: id,
	}, nil
}

func (h *ArticleHandler) Withdraw(ctx *gin.Context, req ArticleWithdrawReq, uc jwt.UserClaims) (ginx.Result, error) {
	err := h.svc.Withdraw(ctx, uc.Uid, req.Id)
	if err != nil {
		return ginx.Result{
			Msg:  "系统错误",
			Code: 5,
		}, err
	}
	return ginx.Result{
		Msg: "OK",
	}, nil
}

func (h *ArticleHandler) Detail(ctx *gin.Context) {
	idstr := ctx.Param("id")
	id, err := strconv.ParseInt(idstr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusOK, ginx.Result{
			Msg:  "id 参数错误",
			Code: 4,
		})
		h.l.Warn("查询文章失败，id格式不对",
			logger.String("id", idstr),
			logger.Error(err))
		return
	}
	art, err := h.svc.GetById(ctx, id)
	if err != nil {
		ctx.JSON(http.StatusOK, ginx.Result{
			Msg:  "系统错误",
			Code: 5,
		})
		h.l.Error("查询文章失败",
			logger.Int64("id", id),
			logger.Error(err))
		return
	}
	uc := ctx.MustGet("user").(jwt.UserClaims)
	if art.Author.Id != uc.Uid {
		// 有人在搞鬼
		ctx.JSON(http.StatusOK, ginx.Result{
			Msg:  "系统错误",
			Code: 5,
		})
		h.l.Error("非法查询",
			logger.Int64("id", id),
			logger.Int64("uid", uc.Uid),
			logger.Error(err))
		return
	}
	vo := ArticleVo{
		Id:    art.Id,
		Title: art.Title,
		//Abstract: art.Abstract(),
		Content:  art.Content,
		AuthorId: art.Author.Id,

		Status: art.Status.ToUint8(),
		Ctime:  art.Ctime.Format(time.DateTime),
		Utime:  art.Utime.Format(time.DateTime),
	}
	ctx.JSON(http.StatusOK, ginx.Result{
		Data: vo,
	})
}

func (h *ArticleHandler) List(ctx *gin.Context) {
	var page Page
	if err := ctx.Bind(&page); err != nil {
		return
	}
	uc := ctx.MustGet("user").(jwt.UserClaims)
	arts, err := h.svc.GetByAuthor(ctx, uc.Uid, page.Offset, page.Limit)
	if err != nil {
		ctx.JSON(http.StatusOK, ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		})
		h.l.Error("查找文章列表失败",
			logger.Error(err),
			logger.Int("offset", page.Offset),
			logger.Int("limit", page.Limit),
			logger.Int64("uid", uc.Uid))
		return
	}
	ctx.JSON(http.StatusOK, ginx.Result{
		Data: slice.Map[domain.Article, ArticleVo](arts, func(idx int, src domain.Article) ArticleVo {
			return ArticleVo{
				Id:       src.Id,
				Title:    src.Title,
				Abstract: src.Abstract(),
				//Content: src.Content,
				// 创作者列表 不需要显示创作者
				Status: src.Status.ToUint8(),
				Ctime:  src.Ctime.Format(time.DateTime),
				Utime:  src.Utime.Format(time.DateTime),
			}
		}),
	})
}

func (h *ArticleHandler) PubDetail(ctx *gin.Context) {
	idstr := ctx.Param("id")
	id, err := strconv.ParseInt(idstr, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusOK, ginx.Result{
			Msg:  "id 参数错误",
			Code: 4,
		})
		h.l.Warn("查询文章失败，id格式不对",
			logger.String("id", idstr),
			logger.Error(err))
		return
	}
	uc := ctx.MustGet("user").(jwt.UserClaims)
	var (
		eg   errgroup.Group
		art  domain.Article
		intr domain.Interactive
	)

	eg.Go(func() error {
		var er error
		art, er = h.svc.GetPubById(ctx, id, uc.Uid)
		return er
	})

	eg.Go(func() error {

		var er error
		intr, er = h.intrSvc.Get(ctx, h.biz, id, uc.Uid)
		return er
	})

	// 等待结果
	err = eg.Wait()
	if err != nil {
		ctx.JSON(http.StatusOK, ginx.Result{
			Msg:  "系统错误",
			Code: 5,
		})
		h.l.Error("查询文章失败，系统错误",
			logger.Int64("aid", id),
			logger.Int64("uid", uc.Uid),
			logger.Error(err))
		return
	}
	//go func() {
	//	// 1. 如果你想摆脱原本主链路的超时控制，你就创建一个新的
	//	// 2. 如果你不想，就直接用ctx
	//	newCtx, cancel := context.WithTimeout(context.Background(), time.Second)
	//	defer cancel()
	//	er := h.intrSvc.IncrReadCnt(newCtx, h.biz, art.Id)
	//	if er != nil {
	//		h.l.Error("更新阅读数失败",
	//			logger.Int64("aid", art.Id),
	//			logger.Error(err))
	//	}
	//}()
	ctx.JSON(http.StatusOK, ginx.Result{
		Data: ArticleVo{
			Id:         art.Id,
			Title:      art.Title,
			Content:    art.Content,
			AuthorId:   art.Author.Id,
			AuthorName: art.Author.Name,
			ReadCnt:    intr.ReadCnt,
			CollectCnt: intr.CollectCnt,
			LikeCnt:    intr.LikeCnt,
			Liked:      intr.Liked,
			Collected:  intr.Collected,

			Status: art.Status.ToUint8(),
			Ctime:  art.Ctime.Format(time.DateTime),
			Utime:  art.Utime.Format(time.DateTime),
		},
	})
}

func (h *ArticleHandler) Like(ctx *gin.Context, req ArticleLikeReq, uc jwt.UserClaims) (ginx.Result, error) {
	var err error
	if req.Like {
		// 点赞
		err = h.intrSvc.Like(ctx, h.biz, req.Id, uc.Uid)
	} else {
		// 取消点赞
		err = h.intrSvc.CancelLike(ctx, h.biz, req.Id, uc.Uid)
	}
	if err != nil {
		return ginx.Result{
			Msg:  "系统错误",
			Code: 5,
		}, err
	}
	return ginx.Result{
		Msg: "OK",
	}, nil
}

func (h *ArticleHandler) Collect(ctx *gin.Context, req ArticleCollectReq, uc jwt.UserClaims) (ginx.Result, error) {
	err := h.intrSvc.Collect(ctx, h.biz, req.Id, req.Cid, uc.Uid)
	if err != nil {
		return ginx.Result{
			Msg:  "系统错误",
			Code: 5,
		}, err
	}
	return ginx.Result{
		Msg: "OK",
	}, nil
}
