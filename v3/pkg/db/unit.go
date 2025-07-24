package db

import (
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"time"
)

type Unit struct {
	DatabaseName   string
	CollectionName string
	Ctx            context.Context
	Client         *mongo.Client
}

type UnitDocument struct {
	Id bson.ObjectID `bson:"_id" json:"-"`

	UnitName       string `bson:"unitName" json:"unitName"`
	ParentUnitName string `bson:"parentUnitName" json:"parentUnitName"`

	Status       string  `bson:"status" json:"status"`
	EntId        string  `bson:"entId" json:"entId"`
	IsBranch     bool    `bson:"isBranch" json:"isBranch"`
	IsInvest     bool    `bson:"isInvest" json:"isInvest"`
	InvestRation float64 `bson:"investRatio" json:"investRatio"`

	CreateTime time.Time `bson:"create_time" json:"-"`
	UpdateTime time.Time `bson:"update_time" json:"-"`
}

func NewUnit(client *mongo.Client) *Unit {
	return &Unit{
		DatabaseName:   GlobalDatabase,
		CollectionName: "unit",
		Ctx:            context.Background(),
		Client:         client,
	}
}

func (i *Unit) Insert(doc UnitDocument) (isSuccess bool, err error) {
	// 生成_id
	if doc.Id.IsZero() {
		doc.Id = bson.NewObjectID()
	}
	now := time.Now()
	doc.CreateTime = now
	doc.UpdateTime = now
	// 插入文档
	col := i.Client.Database(i.DatabaseName).Collection(i.CollectionName)
	_, err = col.InsertOne(i.Ctx, doc)
	if err != nil {
		return
	}
	isSuccess = true
	return
}

func (i *Unit) Update(id bson.ObjectID, update UnitDocument) (isSuccess bool, err error) {
	update.UpdateTime = time.Now()
	// 更新文档
	col := i.Client.Database(i.DatabaseName).Collection(i.CollectionName)
	filter := bson.M{"_id": id}
	updateDoc := bson.M{"$set": update}
	_, err = col.UpdateOne(i.Ctx, filter, updateDoc)
	if err != nil {
		return
	}
	isSuccess = true
	return
}

func (i *Unit) Delete(id string) (isSuccess bool, err error) {
	// 转换id为ObjectId
	idd, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return false, err
	}
	// 删除文档
	col := i.Client.Database(i.DatabaseName).Collection(i.CollectionName)
	filter := bson.M{"_id": idd}
	_, err = col.DeleteOne(i.Ctx, filter)
	if err != nil {
		return
	}
	isSuccess = true
	return
}

func (i *Unit) Get(id string) (doc *UnitDocument, err error) {
	// 转换id为ObjectId
	idd, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	col := i.Client.Database(i.DatabaseName).Collection(i.CollectionName)
	filter := bson.M{"_id": idd}
	var result UnitDocument
	err = col.FindOne(i.Ctx, filter).Decode(&result)
	if err != nil {
		return
	}
	doc = &result
	return
}

func (i *Unit) GetByName(unitName string, parentName string) (doc *UnitDocument, err error) {
	col := i.Client.Database(i.DatabaseName).Collection(i.CollectionName)
	filter := bson.M{"unitName": unitName, "parentUnitName": parentName}
	var result UnitDocument
	err = col.FindOne(i.Ctx, filter).Decode(&result)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, nil
	}
	doc = &result
	return
}

func (i *Unit) InsertOrUpdate(doc UnitDocument) (dss DataSaveStatus, err error) {
	oldDoc, err := i.GetByName(doc.UnitName, doc.ParentUnitName)
	if oldDoc == nil {
		dss.IsNew = true
		dss.IsSuccess, err = i.Insert(doc)
		return
	}
	// 更新文档
	doc.Id = oldDoc.Id
	doc.CreateTime = oldDoc.CreateTime
	doc.UpdateTime = time.Now()

	dss.IsUpdated = true
	dss.IsSuccess, err = i.Update(oldDoc.Id, doc)

	return
}

func (i *Unit) Find(filter bson.M, page, pageSize int) (docs []UnitDocument, err error) {
	col := i.Client.Database(i.DatabaseName).Collection(i.CollectionName)
	opts := options.Find().SetSort(bson.D{{"parentUnitName", 1}, {"unitName", 1}, {UpdateTime, -1}})
	// 计算分页
	if page > 0 && pageSize > 0 {
		opts.SetLimit(int64(pageSize))
		opts.SetSkip(int64((page - 1) * pageSize))
	}
	cursor, err := col.Find(i.Ctx, filter, opts)
	if err != nil {
		return
	}
	defer cursor.Close(i.Ctx)

	if err = cursor.All(i.Ctx, &docs); err != nil {
		return nil, err
	}
	return
}

func (i *Unit) Count(filter bson.M) (int, error) {
	// 计数满足条件的文档
	col := i.Client.Database(i.DatabaseName).Collection(i.CollectionName)
	count, err := col.CountDocuments(i.Ctx, filter)

	return int(count), err
}
