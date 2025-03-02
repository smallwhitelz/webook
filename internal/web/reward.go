package web

import (
	"github.com/gin-gonic/gin"
	rewardv1 "webook/api/proto/gen/reward/v1"
	"webook/internal/web/jwt"
	"webook/pkg/ginx"
	"webook/pkg/logger"
)

type RewardHandler struct {
	client rewardv1.RewardServiceClient
	l      logger.V1
}

func (h *RewardHandler) RegisterRoutes(server *gin.Engine) {
	rg := server.Group("reward")
	rg.Any("/detail", ginx.WrapBodyAndClaims(h.GetReward))
}

type GetRewardReq struct {
	Rid int64
}

func (h *RewardHandler) GetReward(ctx *gin.Context, req GetRewardReq, claims jwt.UserClaims) (ginx.Result, error) {
	resp, err := h.client.GetReward(ctx, &rewardv1.GetRewardRequest{
		// 我这一次打赏的id
		Rid: req.Rid,
		// 防止非法访问,我只能看到我打赏的记录
		// 不能看到别人打赏的记录
		Uid: claims.Uid,
	})
	if err != nil {
		return ginx.Result{
			Code: 5,
			Msg:  "系统错误",
		}, err
	}
	return ginx.Result{
		// 暂时也就是只需要状态
		Data: resp.Status.String(),
	}, nil
}
