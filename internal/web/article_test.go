package web

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"net/http"
	"net/http/httptest"
	"testing"
	"webook/internal/domain"
	"webook/internal/service"
	svcmocks "webook/internal/service/mocks"
	ijwt "webook/internal/web/jwt"
	"webook/pkg/ginx"
	"webook/pkg/logger"
)

func TestArticleHandler_Publish(t *testing.T) {
	testCases := []struct {
		name    string
		mock    func(ctrl *gomock.Controller) (service.ArticleService, service.InteractiveService)
		reqBody string

		wantCode int
		wantRes  ginx.Result
	}{
		{
			name: "新建并发表成功",
			mock: func(ctrl *gomock.Controller) (service.ArticleService, service.InteractiveService) {
				svc := svcmocks.NewMockArticleService(ctrl)
				intrSvc := svcmocks.NewMockInteractiveService(ctrl)
				svc.EXPECT().Publish(gomock.Any(), domain.Article{
					Title:   "我的标题",
					Content: "我的内容",
					Author: domain.Author{
						Id: 123,
					},
				}).Return(int64(1), nil)
				return svc, intrSvc
			},
			reqBody: `
{
    "title": "我的标题",
    "content": "我的内容"
}
`,
			wantCode: http.StatusOK,
			wantRes: ginx.Result{
				Data: float64(1),
			},
		},
		{
			name: "已有帖子发表失败",
			mock: func(ctrl *gomock.Controller) (service.ArticleService, service.InteractiveService) {
				svc := svcmocks.NewMockArticleService(ctrl)
				intrSvc := svcmocks.NewMockInteractiveService(ctrl)
				svc.EXPECT().Publish(gomock.Any(), domain.Article{
					Id:      123,
					Title:   "我的标题",
					Content: "我的内容",
					Author: domain.Author{
						Id: 123,
					},
				}).Return(int64(123), nil)
				return svc, intrSvc
			},
			reqBody: `
{
	"id": 123,
    "title": "我的标题",
    "content": "我的内容"
}
`,
			wantCode: http.StatusOK,
			wantRes: ginx.Result{
				Data: float64(123),
			},
		},
		{
			name: "发表失败",
			mock: func(ctrl *gomock.Controller) (service.ArticleService, service.InteractiveService) {
				svc := svcmocks.NewMockArticleService(ctrl)
				intrSvc := svcmocks.NewMockInteractiveService(ctrl)
				svc.EXPECT().Publish(gomock.Any(), domain.Article{
					Title:   "我的标题",
					Content: "我的内容",
					Author: domain.Author{
						Id: 123,
					},
				}).Return(int64(0), errors.New("发表失败"))
				return svc, intrSvc
			},
			reqBody: `
{
    "title": "我的标题",
    "content": "我的内容"
}
`,
			wantCode: http.StatusOK,
			wantRes: ginx.Result{
				Code: 5,
				Msg:  "系统错误",
			},
		},
		{
			name: "Bind错误",
			mock: func(ctrl *gomock.Controller) (service.ArticleService, service.InteractiveService) {
				svc := svcmocks.NewMockArticleService(ctrl)
				intrSvc := svcmocks.NewMockInteractiveService(ctrl)
				return svc, intrSvc
			},
			reqBody: `
{
    "title": "我的标题",
    "content": "我的内容"asd
}
`,
			wantCode: http.StatusBadRequest,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			// 构造handler
			svc, intrSvc := tc.mock(ctrl)
			hdl := NewArticleHandler(svc, logger.NewNopLogger(), intrSvc)

			// 准备服务器和构造路由
			server := gin.Default()
			server.Use(func(ctx *gin.Context) {
				ctx.Set("user", ijwt.UserClaims{
					Uid: 123,
				})
			})
			hdl.RegisterRoutes(server)

			// 准备Req和记录的 recorder
			req, err := http.NewRequest(http.MethodPost, "/articles/publish",
				bytes.NewBufferString(tc.reqBody))
			req.Header.Set("Content-Type", "application/json")
			assert.NoError(t, err)
			recorder := httptest.NewRecorder()
			// 执行
			server.ServeHTTP(recorder, req)
			assert.Equal(t, tc.wantCode, recorder.Code)
			if tc.wantCode != http.StatusOK {
				return
			}
			// 断言结果
			var res ginx.Result
			err = json.NewDecoder(recorder.Body).Decode(&res)
			assert.NoError(t, err)

			assert.Equal(t, tc.wantRes, res)
		})
	}
}
