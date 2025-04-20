package db

import (
	"context"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"time"
)

type Workspace struct {
	DatabaseName   string
	CollectionName string
	Ctx            context.Context
	Client         *mongo.Client
}

type WorkspaceDocument struct {
	Id bson.ObjectID `bson:"_id"`

	WorkspaceId string   `bson:"workspace_id"`
	Name        string   `bson:"name"`
	Description string   `bson:"description"`
	Status      string   `bson:"status"`
	SortNumber  int      `bson:"sort_number"`
	NotifyId    []string `bson:"notify_id,omitempty"`

	CreateTime time.Time `bson:"create_time" json:"create_time"`
	UpdateTime time.Time `bson:"update_time" json:"update_time"`
}

func NewWorkspace(client *mongo.Client) *Workspace {
	return &Workspace{
		DatabaseName:   GlobalDatabase,
		CollectionName: "workspace",
		Ctx:            context.Background(),
		Client:         client,
	}
}

func (w *Workspace) Insert(doc WorkspaceDocument) (isSuccess bool, err error) {
	// 生成_id
	if doc.Id.IsZero() {
		doc.Id = bson.NewObjectID()
	}
	now := time.Now()
	doc.WorkspaceId = uuid.New().String()
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

func (w *Workspace) Update(id bson.ObjectID, update WorkspaceDocument) (isSuccess bool, err error) {
	now := time.Now()
	update.UpdateTime = now

	// 更新文档
	col := w.Client.Database(w.DatabaseName).Collection(w.CollectionName)
	filter := bson.M{"_id": id}
	updateDoc := bson.M{"$set": update}
	_, err = col.UpdateOne(w.Ctx, filter, updateDoc)
	if err != nil {
		return
	}
	isSuccess = true
	return
}

func (w *Workspace) Delete(id string) (isSuccess bool, err error) {
	// 删除文档
	idd, _ := bson.ObjectIDFromHex(id)
	col := w.Client.Database(w.DatabaseName).Collection(w.CollectionName)
	filter := bson.M{"_id": idd}
	_, err = col.DeleteOne(w.Ctx, filter)
	if err != nil {
		return
	}
	isSuccess = true
	return
}

func (w *Workspace) Find(filter bson.M, page, pageSize int) (docs []WorkspaceDocument, err error) {
	// 查询文档
	col := w.Client.Database(w.DatabaseName).Collection(w.CollectionName)
	opts := options.Find().SetSort(bson.D{{SortNumber, -1}, {"name", 1}, {UpdateTime, -1}})
	// 计算分页
	if page > 0 && pageSize > 0 {
		opts.SetLimit(int64(pageSize))
		opts.SetSkip(int64((page - 1) * pageSize))
	}
	cursor, err := col.Find(w.Ctx, filter, opts)
	if err != nil {
		return
	}
	defer cursor.Close(w.Ctx)

	if err = cursor.All(w.Ctx, &docs); err != nil {
		return nil, err
	}
	return
}

func (w *Workspace) Count(filter bson.M) (int, error) {
	// 计数满足条件的文档
	col := w.Client.Database(w.DatabaseName).Collection(w.CollectionName)
	count, err := col.CountDocuments(w.Ctx, filter)

	return int(count), err
}

func (w *Workspace) Get(id string) (doc WorkspaceDocument, err error) {
	// 查询文档
	idd, _ := bson.ObjectIDFromHex(id)
	col := w.Client.Database(w.DatabaseName).Collection(w.CollectionName)
	filter := bson.M{"_id": idd}
	err = col.FindOne(w.Ctx, filter).Decode(&doc)
	if err != nil {
		return
	}
	return
}

func (w *Workspace) GetByWorkspaceId(workspaceId string) (doc WorkspaceDocument, err error) {
	// 查询文档
	col := w.Client.Database(w.DatabaseName).Collection(w.CollectionName)
	filter := bson.M{"workspace_id": workspaceId}
	err = col.FindOne(w.Ctx, filter).Decode(&doc)
	if err != nil {
		return
	}
	return
}
