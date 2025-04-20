package db

import (
	"context"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"time"
)

type Profile struct {
	DatabaseName   string
	CollectionName string
	Ctx            context.Context
	Client         *mongo.Client
}

type ProfileDocument struct {
	Id bson.ObjectID `bson:"_id" json:"id"`

	ProfileName string `bson:"name" json:"name"`
	Description string `bson:"description" json:"description"`
	Args        string `bson:"args" json:"args"`
	SortNumber  int    `bson:"sort_number" json:"sort_number"`
	Status      string `bson:"status" json:"status"`

	CreateTime time.Time `bson:"create_time"`
	UpdateTime time.Time `bson:"update_time"`
}

func NewProfile(databaseName string, client *mongo.Client) *Profile {
	return &Profile{
		DatabaseName:   databaseName,
		CollectionName: "taskProfile",
		Ctx:            context.Background(),
		Client:         client,
	}
}

func (t *Profile) Insert(doc ProfileDocument) (isSuccess bool, err error) {
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

func (t *Profile) Update(id bson.ObjectID, update ProfileDocument) (isSuccess bool, err error) {
	now := time.Now()
	update.UpdateTime = now

	// 更新文档
	col := t.Client.Database(t.DatabaseName).Collection(t.CollectionName)
	filter := bson.M{"_id": id}
	updateDoc := bson.M{"$set": update}
	_, err = col.UpdateOne(t.Ctx, filter, updateDoc)
	if err != nil {
		return
	}
	isSuccess = true
	return
}

func (t *Profile) Delete(id string) (isSuccess bool, err error) {
	// 删除文档
	idd, _ := bson.ObjectIDFromHex(id)
	col := t.Client.Database(t.DatabaseName).Collection(t.CollectionName)
	filter := bson.M{"_id": idd}
	_, err = col.DeleteOne(t.Ctx, filter)
	if err != nil {
		return
	}
	isSuccess = true
	return
}

func (t *Profile) Find(filter bson.M, page, pageSize int) (docs []ProfileDocument, err error) {
	// 查询文档
	col := t.Client.Database(t.DatabaseName).Collection(t.CollectionName)
	opts := options.Find().SetSort(bson.D{{SortNumber, -1}, {UpdateTime, -1}, {"name", 1}})
	if page > 0 && pageSize > 0 {
		opts.SetSkip(int64((page - 1) * pageSize))
		opts.SetLimit(int64(pageSize))
	}
	cursor, err := col.Find(t.Ctx, filter, opts)
	if err != nil {
		return
	}
	defer cursor.Close(t.Ctx)

	if err = cursor.All(t.Ctx, &docs); err != nil {
		return nil, err
	}
	return
}

func (t *Profile) Count(filter bson.M) (int, error) {
	col := t.Client.Database(t.DatabaseName).Collection(t.CollectionName)
	count, err := col.CountDocuments(t.Ctx, filter)
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

func (t *Profile) Get(id string) (doc ProfileDocument, err error) {
	// 查询文档
	idd, _ := bson.ObjectIDFromHex(id)
	col := t.Client.Database(t.DatabaseName).Collection(t.CollectionName)
	filter := bson.M{"_id": idd}
	err = col.FindOne(t.Ctx, filter).Decode(&doc)
	if err != nil {
		return
	}
	return
}
