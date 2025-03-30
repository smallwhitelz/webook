package dao

type FeedEvent struct {
	id int64
	// 标注一些类型
	Type string
	// 公共字段，比如说的排序字段（推荐度，优先级）
}

type ArticleEvent struct {
	id int64
	// 指向 FeedEvent
	Fid int64
	// 文章的id
	Aid int64
	// 依然可以继续冗余别的字段
	//AuthorName string
	//Title      string
}
