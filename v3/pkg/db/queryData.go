package db

import (
	"context"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"time"
)

type QueryData struct {
	DatabaseName   string
	CollectionName string
	Ctx            context.Context
	Client         *mongo.Client
}

type QueryDocument struct {
	Id         bson.ObjectID `bson:"_id"`
	Category   string        `bson:"category"`
	Domain     string        `bson:"domain"`
	Content    string        `bson:"content"`
	CreateTime time.Time     `bson:"create_time"`
	UpdateTime time.Time     `bson:"update_time"`
}

func NewQueryData(databaseName string, client *mongo.Client) *QueryData {
	return &QueryData{
		DatabaseName:   databaseName,
		CollectionName: "queryData",
		Ctx:            context.Background(),
		Client:         client,
	}
}

func (q *QueryData) Insert(doc *QueryDocument) (isSuccess bool, err error) {
	// 生成_id
	if doc.Id.IsZero() {
		doc.Id = bson.NewObjectID()
	}
	now := time.Now()
	doc.CreateTime = now
	doc.UpdateTime = now
	// 插入文档
	col := q.Client.Database(q.DatabaseName).Collection(q.CollectionName)
	_, err = col.InsertOne(q.Ctx, doc)
	if err != nil {
		return
	}
	isSuccess = true
	return
}

func (q *QueryData) Update(id bson.ObjectID, update QueryDocument) (isSuccess bool, err error) {
	now := time.Now()
	// 更新文档
	col := q.Client.Database(q.DatabaseName).Collection(q.CollectionName)
	filter := bson.M{"_id": id}
	updateDoc := bson.M{"$set": bson.M{"content": update.Content, "update_time": now}}
	_, err = col.UpdateOne(q.Ctx, filter, updateDoc)
	if err != nil {
		return
	}
	isSuccess = true
	return
}

func (q *QueryData) Delete(id bson.ObjectID) (isSuccess bool, err error) {
	// 删除文档
	col := q.Client.Database(q.DatabaseName).Collection(q.CollectionName)
	filter := bson.M{"_id": id}
	_, err = col.DeleteOne(q.Ctx, filter)
	if err != nil {
		return
	}
	isSuccess = true
	return
}

func (q *QueryData) Get(id bson.ObjectID) (doc *QueryDocument, err error) {
	col := q.Client.Database(q.DatabaseName).Collection(q.CollectionName)
	filter := bson.M{"_id": id}
	var result QueryDocument
	err = col.FindOne(q.Ctx, filter).Decode(&result)
	if err != nil {
		return
	}
	doc = &result
	return
}

func (q *QueryData) GetByDomain(domain string, category string) (doc *QueryDocument, err error) {
	col := q.Client.Database(q.DatabaseName).Collection(q.CollectionName)
	filter := bson.M{"domain": domain, QueryCategory: category}
	var result QueryDocument
	err = col.FindOne(q.Ctx, filter).Decode(&result)
	if err != nil {
		return
	}
	doc = &result
	return
}
