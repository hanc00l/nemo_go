package db

import (
	"context"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"time"
)

type CustomData struct {
	DatabaseName   string
	CollectionName string
	Ctx            context.Context
	Client         *mongo.Client
}

type CustomDataDocument struct {
	Id          bson.ObjectID `bson:"_id"`
	Category    string        `bson:"category"`
	Description string        `bson:"description"`
	Data        string        `bson:"data"`
	CreateTime  time.Time     `bson:"create_time"`
	UpdateTime  time.Time     `bson:"update_time"`
}

func NewCustomData(databaseName string, client *mongo.Client) *CustomData {
	return &CustomData{
		DatabaseName:   databaseName,
		CollectionName: "customData",
		Ctx:            context.Background(),
		Client:         client,
	}
}

func (d *CustomData) Insert(doc CustomDataDocument) (isSuccess bool, err error) {
	// 生成_id
	if doc.Id.IsZero() {
		doc.Id = bson.NewObjectID()
	}
	now := time.Now()
	doc.CreateTime = now
	doc.UpdateTime = now
	// 插入文档
	col := d.Client.Database(d.DatabaseName).Collection(d.CollectionName)
	_, err = col.InsertOne(d.Ctx, doc)
	if err != nil {
		return
	}
	isSuccess = true
	return
}

func (d *CustomData) Update(filter bson.M, update bson.M) (isSuccess bool, err error) {
	col := d.Client.Database(d.DatabaseName).Collection(d.CollectionName)
	_, err = col.UpdateOne(d.Ctx, filter, update)
	if err != nil {
		return
	}
	isSuccess = true
	return
}

func (d *CustomData) Delete(id bson.ObjectID) (isSuccess bool, err error) {
	// 删除文档
	col := d.Client.Database(d.DatabaseName).Collection(d.CollectionName)
	filter := bson.M{"_id": id}
	_, err = col.DeleteOne(d.Ctx, filter)
	if err != nil {
		return
	}
	isSuccess = true
	return
}

func (d *CustomData) Get(id string) (doc *CustomData, err error) {
	idd, _ := bson.ObjectIDFromHex(id)
	col := d.Client.Database(d.DatabaseName).Collection(d.CollectionName)
	filter := bson.M{"_id": idd}
	var result CustomData
	err = col.FindOne(d.Ctx, filter).Decode(&result)
	if err != nil {
		return
	}
	doc = &result
	return
}

func (d *CustomData) Find(category string) (result []CustomDataDocument, err error) {
	col := d.Client.Database(d.DatabaseName).Collection(d.CollectionName)
	filter := bson.M{"category": category}
	cur, err := col.Find(d.Ctx, filter)
	// 查询
	if err != nil {
		return
	}
	defer cur.Close(d.Ctx)

	if err = cur.All(d.Ctx, &result); err != nil {
		return nil, err
	}
	return
}
