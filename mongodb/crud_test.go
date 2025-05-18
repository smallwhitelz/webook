package mongodb

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/event"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"testing"
	"time"
)

func TestMongoDB(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	// 为mongodb配置命令监控器
	monitor := &event.CommandMonitor{
		Started: func(ctx context.Context, evt *event.CommandStartedEvent) {
			// 打印出执行的命令
			fmt.Println(evt.Command)
		},
	}
	opts := options.Client().ApplyURI("mongodb://root:root@43.154.97.245:27017").
		SetMonitor(monitor)
	client, err := mongo.Connect(ctx, opts)
	assert.NoError(t, err)
	// 操作 client
	col := client.Database("webook").Collection("articles")
	// 插入一个内容
	insertRes, err := col.InsertOne(ctx, Article{
		Id:       1,
		Title:    "我的标题",
		Content:  "我的内容",
		AuthorId: 123,
	})
	assert.NoError(t, err)
	oid := insertRes.InsertedID.(primitive.ObjectID)
	t.Log("插入ID", oid)
	// 查询内容
	//filter := bson.D{bson.E{"id", 1}}
	filter := bson.M{
		"id": 1,
	}
	findRes := col.FindOne(ctx, filter)
	if findRes.Err() == mongo.ErrNoDocuments {
		t.Log("没找到数据")
	} else {
		assert.NoError(t, findRes.Err())
		var art Article
		err = findRes.Decode(&art)
		assert.NoError(t, err)
		t.Log(art)
	}

	// 修改内容
	updateFilter := bson.D{bson.E{"id", 1}}
	//set := bson.D{bson.E{Key: "$set", Value: bson.E{Key: "title", Value: "新的标题"}}}
	set := bson.D{bson.E{Key: "$set", Value: bson.M{
		"title": "新的标题",
	}}}
	// UpdateOne只会修改第一个id=1的数据
	updateOneRes, err := col.UpdateOne(ctx, updateFilter, set)
	assert.NoError(t, err)
	t.Log("更新文档数量", updateOneRes.ModifiedCount)

	// 记得忽略零值，也就是在字段上加omitempty
	// UpdateMany也是批量修改，这里updateFilter是id=1，那么意味着会修改所有id=1的数据
	updateManyRes, err := col.UpdateMany(ctx, updateFilter,
		bson.D{bson.E{Key: "$set", Value: Article{Content: "新的内容"}}})
	assert.NoError(t, err)
	t.Log("更新文档数量", updateManyRes.ModifiedCount)
	// 删除操作
	deleteFilter := bson.D{bson.E{"id", 1}}
	deleteRes, err := col.DeleteMany(ctx, deleteFilter)
	assert.NoError(t, err)
	t.Log("删除文档数量", deleteRes.DeletedCount)
}

type Article struct {
	Id      int64  `bson:"id,omitempty"`
	Title   string `bson:"title,omitempty"`
	Content string `bson:"content,omitempty"`
	// 我要根据创作者ID来查询
	AuthorId int64 `bson:"author_id,omitempty"`

	Status uint8 `bson:"status,omitempty"`
	// 时区 UTC 0 的毫秒数
	// 创建时间
	Ctime int64 `bson:"ctime,omitempty"`
	// 更新时间
	Utime int64 `bson:"utime,omitempty"`
}
