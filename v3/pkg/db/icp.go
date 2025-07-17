package db

import (
	"context"
	"encoding/json"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"time"
)

type ICP struct {
	DatabaseName   string
	CollectionName string
	Ctx            context.Context
	Client         *mongo.Client
}

type ICPDocument struct {
	Id             bson.ObjectID `bson:"_id" json:"-"`
	Domain         string        `bson:"domain" json:"domain"`
	UnitName       string        `bson:"unitName" json:"unitName"`
	CompanyType    string        `bson:"companyType" json:"companyType"`
	SiteLicense    string        `bson:"siteLicense" json:"siteLicense"`
	ServiceLicence string        `bson:"serviceLicence" json:"serviceLicence"`
	VerifyTime     string        `bson:"verifyTime" json:"verifyTime"`
	Source         string        `bson:"source" json:"source"`
	CreateTime     time.Time     `bson:"create_time" json:"-"`
	UpdateTime     time.Time     `bson:"update_time" json:"-"`
}

func NewICP(client *mongo.Client) *ICP {
	return &ICP{
		DatabaseName:   GlobalDatabase,
		CollectionName: "icp",
		Ctx:            context.Background(),
		Client:         client,
	}
}

func (d *ICPDocument) ToJSONString() string {
	s, err := json.Marshal(d)
	if err != nil {
		return ""
	}
	return string(s)
}

func (i *ICP) Insert(doc ICPDocument) (isSuccess bool, err error) {
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

func (i *ICP) Update(id bson.ObjectID, update ICPDocument) (isSuccess bool, err error) {
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

func (i *ICP) Delete(id bson.ObjectID) (isSuccess bool, err error) {
	// 删除文档
	col := i.Client.Database(i.DatabaseName).Collection(i.CollectionName)
	filter := bson.M{"_id": id}
	_, err = col.DeleteOne(i.Ctx, filter)
	if err != nil {
		return
	}
	isSuccess = true
	return
}

func (i *ICP) Get(id bson.ObjectID) (doc *ICPDocument, err error) {
	col := i.Client.Database(i.DatabaseName).Collection(i.CollectionName)
	filter := bson.M{"_id": id}
	var result ICPDocument
	err = col.FindOne(i.Ctx, filter).Decode(&result)
	if err != nil {
		return
	}
	doc = &result
	return
}

func (i *ICP) GetByDomain(domain string) (doc *ICPDocument, err error) {
	col := i.Client.Database(i.DatabaseName).Collection(i.CollectionName)
	filter := bson.M{"domain": domain}
	var result ICPDocument
	err = col.FindOne(i.Ctx, filter).Decode(&result)
	if err != nil {
		return
	}
	doc = &result
	return
}

func (i *ICP) GetByCompany(company string) (result []ICPDocument, err error) {
	col := i.Client.Database(i.DatabaseName).Collection(i.CollectionName)
	filter := bson.M{"unitName": company}
	cur, err := col.Find(i.Ctx, filter)
	// 查询
	if err != nil {
		return
	}
	defer cur.Close(i.Ctx)

	if err = cur.All(i.Ctx, &result); err != nil {
		return nil, err
	}
	return
}

func (i *ICP) InsertOrUpdate(doc ICPDocument) (dss DataSaveStatus, err error) {
	oldDoc, err := i.GetByDomain(doc.Domain)
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
