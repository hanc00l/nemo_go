package db

import (
	"context"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"time"
)

type MainTask struct {
	DatabaseName   string
	CollectionName string
	Ctx            context.Context
	Client         *mongo.Client
}

type CronTask struct {
	CronEnabled  bool       `bson:"enabled,omitempty"`
	CronExpr     string     `bson:"expr,omitempty"`
	CronRunCount int        `bson:"runCount,omitempty"`
	CronLastRun  *time.Time `bson:"lastRun,omitempty"`
}

type MainTaskDocument struct {
	Id          bson.ObjectID `bson:"_id"`
	WorkspaceId string        `bson:"workspaceId"`

	TaskId        string `bson:"taskId"`
	TaskName      string `bson:"name"`
	Description   string `bson:"description"`
	Target        string `bson:"target"`
	ExcludeTarget string `bson:"excludeTarget"`
	ProfileName   string `bson:"profileName"`
	OrgId         string `bson:"orgId"`
	IsProxy       bool   `bson:"proxy"`
	Args          string `bson:"args"`

	TargetSliceType int `bson:"targetSliceType"`
	TargetSliceNum  int `bson:"targetSliceNum,omitempty"`

	IsCron       bool      `bson:"cron"`
	CronTaskInfo *CronTask `bson:"cronTaskInfo,omitempty"`

	Status       string     `bson:"status"`
	Result       string     `bson:"result,omitempty"`
	Progress     string     `bson:"progress,omitempty"`
	ProgressRate float64    `bson:"progressRate,omitempty"`
	StartTime    *time.Time `bson:"start_time,omitempty"`
	EndTime      *time.Time `bson:"end_time,omitempty"`

	NotifyId []string `bson:"notifyId,omitempty"`

	CreateTime time.Time `bson:"create_time"`
	UpdateTime time.Time `bson:"update_time"`
}

func NewMainTask(client *mongo.Client) *MainTask {
	return &MainTask{
		DatabaseName:   GlobalDatabase,
		CollectionName: "mainTask",
		Ctx:            context.Background(),
		Client:         client,
	}
}

func (t *MainTask) Insert(doc MainTaskDocument) (isSuccess bool, err error) {
	// 生成_id
	if doc.Id.IsZero() {
		doc.Id = bson.NewObjectID()
	}
	now := time.Now()
	doc.CreateTime = now
	doc.UpdateTime = now
	// 插入文档
	col := t.Client.Database(t.DatabaseName).Collection(t.CollectionName)
	_, err = col.InsertOne(t.Ctx, doc)
	if err != nil {
		return
	}
	isSuccess = true
	return
}

func (t *MainTask) GetById(id string) (result MainTaskDocument, err error) {
	// 查询文档
	col := t.Client.Database(t.DatabaseName).Collection(t.CollectionName)
	idd, _ := bson.ObjectIDFromHex(id)
	filter := bson.M{ID: idd}
	err = col.FindOne(t.Ctx, filter).Decode(&result)
	if err != nil {
		return
	}
	return
}

func (t *MainTask) GetByTaskId(taskId string) (result MainTaskDocument, err error) {
	// 查询文档
	col := t.Client.Database(t.DatabaseName).Collection(t.CollectionName)
	filter := bson.M{TaskId: taskId}
	err = col.FindOne(t.Ctx, filter).Decode(&result)
	if err != nil {
		return
	}
	return
}

func (t *MainTask) Find(filter bson.M, page, pageSize int) (result []MainTaskDocument, err error) {
	// 查询文档
	col := t.Client.Database(t.DatabaseName).Collection(t.CollectionName)
	// 计算分页
	opts := options.Find()
	if pageSize > 0 {
		opts.SetLimit(int64(pageSize))
		opts.SetSkip(int64((page - 1) * pageSize))
	}
	opts.SetSort(bson.M{UpdateTime: -1})
	cur, err := col.Find(t.Ctx, filter, opts)
	// 查询
	if err != nil {
		return
	}
	defer cur.Close(t.Ctx)

	if err = cur.All(t.Ctx, &result); err != nil {
		return nil, err
	}
	return
}

func (t *MainTask) Count(filter bson.M) (int, error) {
	col := t.Client.Database(t.DatabaseName).Collection(t.CollectionName)
	count, err := col.CountDocuments(t.Ctx, filter)
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

func (t *MainTask) Update(id string, update bson.M) (isSuccess bool, err error) {
	// 更新文档
	idd, _ := bson.ObjectIDFromHex(id)
	filter := bson.M{ID: idd}
	update[UpdateTime] = time.Now()
	updateDoc := bson.M{"$set": update}
	col := t.Client.Database(t.DatabaseName).Collection(t.CollectionName)
	result, err := col.UpdateOne(t.Ctx, filter, updateDoc)
	if err != nil {
		return false, err
	}
	if result.ModifiedCount > 0 {
		isSuccess = true
	}
	return
}

func (t *MainTask) Delete(filter bson.M) (isSuccess bool, err error) {
	// 删除文档
	col := t.Client.Database(t.DatabaseName).Collection(t.CollectionName)
	result, err := col.DeleteOne(t.Ctx, filter)
	if err != nil {
		return false, err
	}
	if result.DeletedCount > 0 {
		isSuccess = true
	}
	return
}
