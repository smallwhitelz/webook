package service

import (
	"context"
	"github.com/ecodeclub/ekit/queue"
	"github.com/ecodeclub/ekit/slice"
	"math"
	"time"
	"webook/internal/domain"
	"webook/internal/repository"
)

//go:generate mockgen -source=./ranking_service.go -package=svcmocks -destination=./mocks/ranking_service.mock.go RankingService
type RankingService interface {
	// TopN 前 100 的
	TopN(ctx context.Context) error
	GetTopN(ctx context.Context) ([]domain.Article, error)
}

type BatchRankingService struct {
	// 用来取点赞数
	intrSvc InteractiveService

	// 用来查找文章
	artSvc ArticleService

	// 控制批量获取的大小
	batchSize int

	// 数学计算文章分数，这里在初始化就搞好，只关心算法而不关心数学计算
	scoreFunc func(likeCnt int64, utime time.Time) float64

	// 前n篇文章
	n int

	repo repository.RankingRepository
}

func (b *BatchRankingService) GetTopN(ctx context.Context) ([]domain.Article, error) {
	return b.repo.GetTopN(ctx)
}

func NewBatchRankingService(intrSvc InteractiveService, artSvc ArticleService, repo repository.RankingRepository) RankingService {
	return &BatchRankingService{
		intrSvc:   intrSvc,
		artSvc:    artSvc,
		batchSize: 100,
		n:         100,
		repo:      repo,
		scoreFunc: func(likeCnt int64, utime time.Time) float64 {
			// 时间
			duration := time.Since(utime).Seconds()
			return float64(likeCnt-1) / math.Pow(duration+2, 1.5)
		},
	}
}

func (b *BatchRankingService) TopN(ctx context.Context) error {
	arts, err := b.topN(ctx)
	if err != nil {
		return err
	}
	// 最终是要放到缓存里面的
	// 存到缓存里
	return b.repo.ReplaceTopN(ctx, arts)
}

func (b *BatchRankingService) topN(ctx context.Context) ([]domain.Article, error) {
	offset := 0
	start := time.Now()
	// 7天前的数据必不可能是热榜，所以没必要取7天前的数据
	ddl := start.Add(-7 * 24 * time.Hour)
	type Score struct {
		score float64
		art   domain.Article
	}
	// NewPriorityQueue 小顶堆的优先级队列
	topN := queue.NewPriorityQueue(b.n, func(src Score, dst Score) int {
		if src.score > dst.score {
			return 1
		} else if src.score == dst.score {
			return 0
		} else {
			return -1
		}
	})
	for {
		// 取数据
		arts, err := b.artSvc.ListPub(ctx, start, offset, b.batchSize)
		if err != nil {
			return nil, err
		}
		// 如果取出来没有文章了，说明取完了，就没必要往下继续执行
		if len(arts) == 0 {
			break
		}
		ids := slice.Map[domain.Article, int64](arts, func(idx int, art domain.Article) int64 {
			return art.Id
		})
		// 取点赞数
		intrMap, err := b.intrSvc.GetByIds(ctx, "article", ids)
		if err != nil {
			return nil, err
		}
		for _, art := range arts {
			intr := intrMap[art.Id]
			score := b.scoreFunc(intr.LikeCnt, art.Utime)
			ele := Score{
				score: score,
				art:   art,
			}
			err = topN.Enqueue(ele)
			if err == queue.ErrOutOfCapacity {
				// 这个也是满了
				// 拿出最小的元素
				// 如果这里用Peek方法，就不用说是再放回去，因为Peek就是看一眼
				minEle, _ := topN.Dequeue()
				if minEle.score < score {
					_ = topN.Enqueue(ele)
				} else {
					// 否则就将取出来的元素放回去，
					_ = topN.Enqueue(minEle)
				}
			}
		}
		offset = offset + len(arts)
		// 没有取够一批，我们就直接中断执行
		// 没有下一批了
		// 最后篇文章的时间如果早于7天前，直接break，没必要继续取
		if len(arts) < b.batchSize || arts[len(arts)-1].Utime.Before(ddl) {
			break
		}
	}
	// 这边topN里面就是最终结果
	res := make([]domain.Article, topN.Len())
	// 因为是小顶堆，Dequeue取出来是最小的元素，所以要把最小的放到最后的索引哪里
	for i := topN.Len() - 1; i >= 0; i-- {
		ele, _ := topN.Dequeue()
		res[i] = ele.art
	}
	return res, nil
}
