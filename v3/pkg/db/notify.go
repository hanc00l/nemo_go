package db

import (
	"context"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"time"
)

type Notify struct {
	DatabaseName   string
	CollectionName string
	Ctx            context.Context
	Client         *mongo.Client
}

type NotifyDocument struct {
	Id bson.ObjectID `bson:"_id"`

	Name        string    `bson:"name"`
	Description string    `bson:"description"`
	Category    string    `bson:"category"`
	Template    string    `bson:"template"`
	Token       string    `bson:"token"`
	SortNumber  int       `bson:"sort_number"`
	Status      string    `bson:"status"`
	CreateTime  time.Time `bson:"create_time"`
	UpdateTime  time.Time `bson:"update_time"`
}

func NewNotify(client *mongo.Client) *Notify {
	return &Notify{
		DatabaseName:   GlobalDatabase,
		CollectionName: "notify",
		Ctx:            context.Background(),
		Client:         client,
	}
}

func (n *Notify) Insert(doc NotifyDocument) (isSuccess bool, err error) {
	// 生成_id
	if doc.Id.IsZero() {
		doc.Id = bson.NewObjectID()
	}
	now := time.Now()
	doc.CreateTime = now
	doc.UpdateTime = now

	// 插入文档
	col := n.Client.Database(n.DatabaseName).Collection(n.CollectionName)
	_, err = col.InsertOne(n.Ctx, doc)
	if err != nil {
		return
	}
	isSuccess = true
	return
}

func (n *Notify) Get(id string) (doc NotifyDocument, err error) {
	// 查询文档
	idd, _ := bson.ObjectIDFromHex(id)
	col := n.Client.Database(n.DatabaseName).Collection(n.CollectionName)
	filter := bson.M{"_id": idd}
	err = col.FindOne(n.Ctx, filter).Decode(&doc)
	if err != nil {
		return
	}
	return
}

func (n *Notify) Find(filter bson.M, page, pageSize int) (docs []NotifyDocument, err error) {
	col := n.Client.Database(n.DatabaseName).Collection(n.CollectionName)
	opts := options.Find().SetSort(bson.D{{SortNumber, -1}, {"name", 1}, {UpdateTime, -1}})
	// 计算分页
	if page > 0 && pageSize > 0 {
		opts.SetLimit(int64(pageSize))
		opts.SetSkip(int64((page - 1) * pageSize))
	}
	cursor, err := col.Find(n.Ctx, filter, opts)
	if err != nil {
		return
	}
	defer cursor.Close(n.Ctx)

	if err = cursor.All(n.Ctx, &docs); err != nil {
		return nil, err
	}
	return
}

func (n *Notify) Count(filter bson.M) (int, error) {
	// 计数满足条件的文档
	col := n.Client.Database(n.DatabaseName).Collection(n.CollectionName)
	count, err := col.CountDocuments(n.Ctx, filter)

	return int(count), err
}

func (n *Notify) Update(id string, update NotifyDocument) (isSuccess bool, err error) {
	now := time.Now()
	update.UpdateTime = now

	// 更新文档
	idd, _ := bson.ObjectIDFromHex(id)
	col := n.Client.Database(n.DatabaseName).Collection(n.CollectionName)
	filter := bson.M{"_id": idd}
	updateDoc := bson.M{"$set": update}
	_, err = col.UpdateOne(n.Ctx, filter, updateDoc)
	if err != nil {
		return
	}
	isSuccess = true
	return
}

func (n *Notify) Delete(id string) (isSuccess bool, err error) {
	// 删除文档
	idd, _ := bson.ObjectIDFromHex(id)
	col := n.Client.Database(n.DatabaseName).Collection(n.CollectionName)
	filter := bson.M{"_id": idd}
	_, err = col.DeleteOne(n.Ctx, filter)
	if err != nil {
		return
	}
	isSuccess = true
	return
}
