package db

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"time"
)

type Vul struct {
	DatabaseName   string
	CollectionName string
	Client         *mongo.Client
	Ctx            context.Context
	TaskId         string // 任务ID：用于区分不同任务的历史记录；为空时表示是总的历史库记录
}

type VulDocument struct {
	Id bson.ObjectID `bson:"_id" json:"-"`

	Authority string `bson:"authority" json:"authority"`
	Host      string `bson:"host" json:"host"`
	Url       string `bson:"url" json:"url"`
	Source    string `bson:"source" json:"source"`
	PocFile   string `bson:"pocfile" json:"pocfile"`
	Name      string `bson:"name" json:"name"`
	Severity  string `bson:"severity" json:"severity"`
	Extra     string `bson:"extra,omitempty" json:"extra,omitempty"`
	TaskId    string `bson:"taskId" json:"taskId,omitempty"`

	CreateTime time.Time `bson:"create_time" json:"-"`
	UpdateTime time.Time `bson:"update_time" json:"-"`
}

func NewVul(databaseName string, collectionName string, client *mongo.Client) *Vul {
	vul := &Vul{
		DatabaseName:   databaseName,
		CollectionName: collectionName,
		Ctx:            context.Background(),
		Client:         client,
	}

	return vul
}

func (v *VulDocument) Equal(other *VulDocument) bool {
	if v.Authority != other.Authority || v.Url != other.Url || v.Source != other.Source || v.PocFile != other.PocFile || v.TaskId != other.TaskId {
		return false
	}
	return true
}

func (w *Vul) Insert(doc VulDocument) (isSuccess bool, err error) {
	// 生成_id
	if doc.Id.IsZero() {
		doc.Id = bson.NewObjectID()
	}
	now := time.Now()
	doc.CreateTime = now
	doc.UpdateTime = now

	// 插入文档
	col := w.Client.Database(w.DatabaseName).Collection(w.CollectionName)
	_, err = col.InsertOne(w.Ctx, doc)
	if err != nil {
		return
	}
	isSuccess = true
	return
}

func (w *Vul) Update(id string, update VulDocument) (isSuccess bool, err error) {
	idd, _ := bson.ObjectIDFromHex(id)
	now := time.Now()
	update.UpdateTime = now

	// 更新文档
	col := w.Client.Database(w.DatabaseName).Collection(w.CollectionName)
	filter := bson.M{"_id": idd}
	updateDoc := bson.M{"$set": update}
	_, err = col.UpdateOne(w.Ctx, filter, updateDoc)
	if err != nil {
		return
	}
	isSuccess = true
	return
}

func (w *Vul) UpdateTime(id string) (isSuccess bool, err error) {
	idd, _ := bson.ObjectIDFromHex(id)

	// 更新文档
	col := w.Client.Database(w.DatabaseName).Collection(w.CollectionName)
	filter := bson.M{"_id": idd}
	updateDoc := bson.M{"$set": bson.M{UpdateTime: time.Now()}}
	_, err = col.UpdateOne(w.Ctx, filter, updateDoc)
	if err != nil {
		return
	}
	isSuccess = true
	return
}

func (w *Vul) Delete(id string) (isSuccess bool, err error) {
	idd, _ := bson.ObjectIDFromHex(id)
	// 删除文档
	col := w.Client.Database(w.DatabaseName).Collection(w.CollectionName)
	filter := bson.M{"_id": idd}
	_, err = col.DeleteOne(w.Ctx, filter)
	if err != nil {
		return
	}
	isSuccess = true
	return
}

func (w *Vul) Find(filter bson.M, page, pageSize int) (docs []VulDocument, err error) {
	// 查询文档
	col := w.Client.Database(w.DatabaseName).Collection(w.CollectionName)
	// 计算分页
	opts := options.Find()
	if pageSize > 0 {
		opts.SetLimit(int64(pageSize))
		opts.SetSkip(int64((page - 1) * pageSize))
	}
	opts.SetSort(bson.D{{UpdateTime, -1}, {"authority", 1}})
	cur, err := col.Find(w.Ctx, filter, opts)
	if err != nil {
		return
	}
	defer cur.Close(w.Ctx)

	if err = cur.All(w.Ctx, &docs); err != nil {
		return nil, err
	}
	return
}

func (w *Vul) Count(filter bson.M) (int, error) {
	col := w.Client.Database(w.DatabaseName).Collection(w.CollectionName)
	count, err := col.CountDocuments(w.Ctx, filter)
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

func (w *Vul) Get(id string) (doc VulDocument, err error) {
	idd, _ := bson.ObjectIDFromHex(id)
	// 查询文档
	col := w.Client.Database(w.DatabaseName).Collection(w.CollectionName)
	filter := bson.M{"_id": idd}
	err = col.FindOne(w.Ctx, filter).Decode(&doc)
	if err != nil {
		return
	}
	return
}

func (w *Vul) Aggregate(field string, limit int) (results []StatisticData, err error) {
	// 获取集合
	collection := w.Client.Database(w.DatabaseName).Collection(w.CollectionName)
	// 定义聚合管道
	pipeline := []bson.M{
		{
			"$group": bson.M{
				"_id":   "$" + field,
				"count": bson.M{"$sum": 1},
			},
		},
	}
	// 排序和限制
	pipeline = append(pipeline, bson.M{"$sort": bson.M{"count": -1}})
	if limit > 0 {
		pipeline = append(pipeline, bson.M{"$limit": limit})
	}

	// 执行聚合查询
	cursor, err := collection.Aggregate(w.Ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(w.Ctx)

	// 解析结果为 StatisticData 数组
	err = cursor.All(w.Ctx, &results)
	return
}

// GetDailyStats 获取最近N天的统计信息
func (w *Vul) GetDailyStats(days int) (ApexChartData, error) {
	collection := w.Client.Database(w.DatabaseName).Collection(w.CollectionName)
	// 1. 计算日期范围
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -days+1) // 包含今天

	// 2. 执行聚合查询
	pipeline := mongo.Pipeline{
		// 匹配指定日期范围内的记录
		{{"$match", bson.D{
			{"update_time", bson.D{
				{"$gte", startDate},
				{"$lte", endDate},
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

// generateChartData 生成ApexCharts所需数据结构
func generateChartData(startDate, endDate time.Time, stats []DailyStat) ApexChartData {
	// 创建日期到计数的映射
	statMap := make(map[string]int)
	for _, s := range stats {
		statMap[s.ID.Day] = s.Count
	}

	// 生成完整日期范围
	var categories []string
	var data []int

	// 遍历每一天
	for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
		dateStr := d.Format("2006-01-02")
		dayOnly := d.Format("02") // 只取"日"部分

		categories = append(categories, dayOnly)

		if count, exists := statMap[dateStr]; exists {
			data = append(data, count)
		} else {
			data = append(data, 0) // 没有数据的日期填充0
		}
	}

	// 构建ApexCharts数据结构
	return ApexChartData{
		Categories: categories,
		Series: []ApexChartSeries{
			{
				Name: "每日记录数",
				Data: data,
			},
		},
	}
}
