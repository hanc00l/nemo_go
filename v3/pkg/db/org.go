package db

import (
	"context"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"time"
)

type Org struct {
	DatabaseName   string
	CollectionName string
	Client         *mongo.Client
	Ctx            context.Context
}

type OrgDocument struct {
	Id          bson.ObjectID `bson:"_id"`
	Name        string        `bson:"name"`
	Description string        `bson:"description"`
	SortNumber  int           `bson:"sort_number"`
	Status      string        `bson:"status"`
	CreateTime  time.Time     `bson:"create_time"`
	UpdateTime  time.Time     `bson:"update_time"`
}

func NewOrg(databaseName string, client *mongo.Client) *Org {
	return &Org{
		DatabaseName:   databaseName,
		CollectionName: "org",
		Client:         client,
		Ctx:            context.Background(),
	}
}

func (o *Org) Insert(doc OrgDocument) (isSuccess bool, err error) {
	// 生成_id
	if doc.Id.IsZero() {
		doc.Id = bson.NewObjectID()
	}
	now := time.Now()
	doc.CreateTime = now
	doc.UpdateTime = now

	// 插入文档
	col := o.Client.Database(o.DatabaseName).Collection(o.CollectionName)
	_, err = col.InsertOne(o.Ctx, doc)
	if err != nil {
		return
	}
	isSuccess = true
	return
}

func (o *Org) Get(id string) (doc OrgDocument, err error) {
	// 查询文档
	idd, _ := bson.ObjectIDFromHex(id)
	col := o.Client.Database(o.DatabaseName).Collection(o.CollectionName)
	filter := bson.M{"_id": idd}
	err = col.FindOne(o.Ctx, filter).Decode(&doc)
	if err != nil {
		return
	}
	return
}

func (o *Org) GetByName(orgName string) (doc OrgDocument, err error) {
	// 查询文档
	col := o.Client.Database(o.DatabaseName).Collection(o.CollectionName)
	filter := bson.M{"name": orgName}
	err = col.FindOne(o.Ctx, filter).Decode(&doc)
	if err != nil {
		return
	}
	return
}

func (o *Org) Find(filter bson.M, page, pageSize int) (docs []OrgDocument, err error) {
	col := o.Client.Database(o.DatabaseName).Collection(o.CollectionName)
	opts := options.Find().SetSort(bson.D{{SortNumber, -1}, {"name", 1}, {UpdateTime, -1}})
	// 计算分页
	if page > 0 && pageSize > 0 {
		opts.SetLimit(int64(pageSize))
		opts.SetSkip(int64((page - 1) * pageSize))
	}
	cursor, err := col.Find(o.Ctx, filter, opts)
	if err != nil {
		return
	}
	defer cursor.Close(o.Ctx)

	if err = cursor.All(o.Ctx, &docs); err != nil {
		return nil, err
	}
	return
}

func (o *Org) Count(filter bson.M) (int, error) {
	// 计数满足条件的文档
	col := o.Client.Database(o.DatabaseName).Collection(o.CollectionName)
	count, err := col.CountDocuments(o.Ctx, filter)

	return int(count), err
}

func (o *Org) Update(id string, update OrgDocument) (isSuccess bool, err error) {
	now := time.Now()
	update.UpdateTime = now

	// 更新文档
	idd, _ := bson.ObjectIDFromHex(id)
	col := o.Client.Database(o.DatabaseName).Collection(o.CollectionName)
	filter := bson.M{"_id": idd}
	updateDoc := bson.M{"$set": update}
	_, err = col.UpdateOne(o.Ctx, filter, updateDoc)
	if err != nil {
		return
	}
	isSuccess = true
	return
}

func (o *Org) Delete(id string) (isSuccess bool, err error) {
	// 删除文档
	idd, _ := bson.ObjectIDFromHex(id)
	col := o.Client.Database(o.DatabaseName).Collection(o.CollectionName)
	filter := bson.M{"_id": idd}
	_, err = col.DeleteOne(o.Ctx, filter)
	if err != nil {
		return
	}
	isSuccess = true
	return
}
