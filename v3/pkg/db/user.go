package db

import (
	"context"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"time"
)

type User struct {
	DatabaseName   string
	CollectionName string
	Ctx            context.Context
	Client         *mongo.Client
}

type UserDocument struct {
	Id bson.ObjectID `bson:"_id"`

	Username    string   `bson:"username"`
	Password    string   `bson:"password"`
	Description string   `bson:"description"`
	Role        string   `bson:"role"`
	SortNumber  int      `bson:"sort_number"`
	Status      string   `bson:"status"`
	WorkspaceId []string `bson:"workspace_id"`

	CreateTime time.Time `bson:"create_time" json:"create_time"`
	UpdateTime time.Time `bson:"update_time" json:"update_time"`
}

func NewUser(client *mongo.Client) *User {
	return &User{
		DatabaseName:   GlobalDatabase,
		CollectionName: "user",
		Ctx:            context.Background(),
		Client:         client,
	}
}

func (u *User) Insert(doc UserDocument) (isSuccess bool, err error) {
	// 生成_id
	if doc.Id.IsZero() {
		doc.Id = bson.NewObjectID()
	}
	now := time.Now()
	doc.CreateTime = now
	doc.UpdateTime = now

	// 插入文档
	col := u.Client.Database(u.DatabaseName).Collection(u.CollectionName)
	_, err = col.InsertOne(u.Ctx, doc)
	if err != nil {
		return
	}
	isSuccess = true
	return
}

func (u *User) Update(id string, update UserDocument) (isSuccess bool, err error) {
	now := time.Now()
	update.UpdateTime = now

	// 更新文档
	_idd, _ := bson.ObjectIDFromHex(id)
	col := u.Client.Database(u.DatabaseName).Collection(u.CollectionName)
	filter := bson.M{"_id": _idd}
	updateDoc := bson.M{"$set": update}
	_, err = col.UpdateOne(u.Ctx, filter, updateDoc)
	if err != nil {
		return
	}
	isSuccess = true
	return
}

func (u *User) Delete(id string) (isSuccess bool, err error) {
	// 删除文档
	idd, _ := bson.ObjectIDFromHex(id)
	col := u.Client.Database(u.DatabaseName).Collection(u.CollectionName)
	filter := bson.M{"_id": idd}
	_, err = col.DeleteOne(u.Ctx, filter)
	if err != nil {
		return
	}
	isSuccess = true
	return
}

func (u *User) Find(filer bson.M, page, pageSize int) (docs []UserDocument, err error) {
	// 查询文档
	col := u.Client.Database(u.DatabaseName).Collection(u.CollectionName)
	opts := options.Find().SetSort(bson.D{{SortNumber, -1}, {"username", 1}, {UpdateTime, -1}})
	// 计算分页
	if page > 0 && pageSize > 0 {
		opts.SetLimit(int64(pageSize))
		opts.SetSkip(int64((page - 1) * pageSize))
	}
	cursor, err := col.Find(u.Ctx, filer, opts)
	if err != nil {
		return
	}
	defer cursor.Close(u.Ctx)

	if err = cursor.All(u.Ctx, &docs); err != nil {
		return nil, err
	}
	return
}

func (u *User) Count(filer bson.M) (int, error) {
	// 计数满足条件的文档
	col := u.Client.Database(u.DatabaseName).Collection(u.CollectionName)
	count, err := col.CountDocuments(u.Ctx, filer)

	return int(count), err
}

func (u *User) Get(id string) (doc UserDocument, err error) {
	// 查询文档
	idd, _ := bson.ObjectIDFromHex(id)
	col := u.Client.Database(u.DatabaseName).Collection(u.CollectionName)
	filter := bson.M{"_id": idd}
	err = col.FindOne(u.Ctx, filter).Decode(&doc)
	if err != nil {
		return
	}
	return
}

func (u *User) GetByName(name string) (doc UserDocument, err error) {
	col := u.Client.Database(u.DatabaseName).Collection(u.CollectionName)
	filter := bson.M{"username": name}
	err = col.FindOne(u.Ctx, filter).Decode(&doc)
	if err != nil {
		return
	}
	return
}
