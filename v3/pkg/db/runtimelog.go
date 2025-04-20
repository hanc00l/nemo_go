package db

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"time"
)

type RuntimeLog struct {
	DatabaseName   string
	CollectionName string
	Ctx            context.Context
	Client         *mongo.Client
}

type RuntimeLogDocument struct {
	Id         bson.ObjectID `bson:"_id,omitempty"`
	Source     string        `bson:"source"`
	File       string        `bson:"file"`
	Func       string        `bson:"func"`
	Level      string        `bson:"level"`
	LevelInt   int           `bson:"level_int"`
	Message    string        `bson:"message"`
	CreateTime time.Time     `bson:"create_time"`
	UpdateTime time.Time     `bson:"update_time"`
}

func NewRuntimeLog(client *mongo.Client) *RuntimeLog {
	return &RuntimeLog{
		DatabaseName:   GlobalDatabase,
		CollectionName: "runtimelog",
		Ctx:            context.Background(),
		Client:         client,
	}
}

func (l *RuntimeLog) Insert(doc RuntimeLogDocument) (isSuccess bool, err error) {
	// 生成_id
	if doc.Id.IsZero() {
		doc.Id = bson.NewObjectID()
	}
	switch doc.Level {
	case "panic":
		doc.LevelInt = int(logrus.PanicLevel)
	case "fatal":
		doc.LevelInt = int(logrus.FatalLevel)
	case "error":
		doc.LevelInt = int(logrus.ErrorLevel)
	case "warning":
		doc.LevelInt = int(logrus.WarnLevel)
	case "info":
		doc.LevelInt = int(logrus.InfoLevel)
	case "debug":
		doc.LevelInt = int(logrus.DebugLevel)
	default:
		doc.LevelInt = 10
	}
	now := time.Now()
	doc.CreateTime = now
	doc.UpdateTime = now

	// 插入文档
	col := l.Client.Database(l.DatabaseName).Collection(l.CollectionName)
	_, err = col.InsertOne(l.Ctx, doc)
	if err != nil {
		return
	}
	isSuccess = true
	return
}

func (l *RuntimeLog) Get(id string) (doc RuntimeLogDocument, err error) {
	// 查询文档
	idd, _ := bson.ObjectIDFromHex(id)
	col := l.Client.Database(l.DatabaseName).Collection(l.CollectionName)
	filter := bson.M{"_id": idd}
	err = col.FindOne(l.Ctx, filter).Decode(&doc)
	if err != nil {
		return
	}
	return
}

func (l *RuntimeLog) Find(filter bson.M, page, pageSize int) (docs []RuntimeLogDocument, err error) {
	col := l.Client.Database(l.DatabaseName).Collection(l.CollectionName)
	opts := options.Find().SetSort(bson.D{{UpdateTime, -1}})
	// 计算分页
	if page > 0 && pageSize > 0 {
		opts.SetLimit(int64(pageSize))
		opts.SetSkip(int64((page - 1) * pageSize))
	}
	cursor, err := col.Find(l.Ctx, filter, opts)
	if err != nil {
		return
	}
	defer cursor.Close(l.Ctx)

	if err = cursor.All(l.Ctx, &docs); err != nil {
		return nil, err
	}
	return
}

func (l *RuntimeLog) Count(filter bson.M) (int, error) {
	// 计数满足条件的文档
	col := l.Client.Database(l.DatabaseName).Collection(l.CollectionName)
	count, err := col.CountDocuments(l.Ctx, filter)

	return int(count), err
}

func (l *RuntimeLog) Delete(id string) (isSuccess bool, err error) {
	// 删除文档
	idd, _ := bson.ObjectIDFromHex(id)
	col := l.Client.Database(l.DatabaseName).Collection(l.CollectionName)
	filter := bson.M{"_id": idd}
	_, err = col.DeleteOne(l.Ctx, filter)
	if err != nil {
		return
	}
	isSuccess = true
	return
}

func (l *RuntimeLog) DeleteMany(filter bson.M) (isSuccess bool, err error) {
	// 删除文档
	col := l.Client.Database(l.DatabaseName).Collection(l.CollectionName)
	_, err = col.DeleteMany(l.Ctx, filter)
	if err != nil {
		return
	}
	isSuccess = true
	return
}

// GetLoginStats 获取最近N天的登录统计信息
func (l *RuntimeLog) GetLoginStats(days int) (ApexChartData, error) {
	collection := l.Client.Database(l.DatabaseName).Collection(l.CollectionName)

	// 1. 计算日期范围
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -days+1) // 包含今天

	// 2. 执行聚合查询
	pipeline := mongo.Pipeline{
		// 匹配指定日期范围和条件的记录
		{{"$match", bson.D{
			{"update_time", bson.D{
				{"$gte", startDate},
				{"$lte", endDate},
			}},
			{"func", "controllers.(*LoginController).LoginAction"},
			{"message", bson.D{
				{"$regex", "login from"},
			}},
		}}},
		// 按天分组统计
		{{"$group", bson.D{
			{"_id", bson.D{
				{"day", bson.D{
					{"$dateToString", bson.D{
						{"format", "%Y-%m-%d"},
						{"date", "$update_time"},
					}},
				}},
			}},
			{"count", bson.D{{"$sum", 1}}},
		}}},
		// 按日期排序
		{{"$sort", bson.D{
			{"_id.day", 1},
		}}},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return ApexChartData{}, fmt.Errorf("aggregation failed: %v", err)
	}
	defer cursor.Close(ctx)

	// 3. 处理聚合结果
	var dailyStats []DailyStat
	if err := cursor.All(ctx, &dailyStats); err != nil {
		return ApexChartData{}, fmt.Errorf("cursor decode failed: %v", err)
	}

	// 4. 生成完整日期范围并填充数据
	return generateChartData(startDate, endDate, dailyStats), nil
}
