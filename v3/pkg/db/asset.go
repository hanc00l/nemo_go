package db

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"github.com/hanc00l/nemo_go/v3/pkg/conf"
	"github.com/hanc00l/nemo_go/v3/pkg/utils"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
)

type RecordToDocFunc func(headers, record []string, orgId string) (*AssetDocument, error)
type Asset struct {
	DatabaseName   string
	CollectionName string
	Ctx            context.Context
	Client         *mongo.Client
	TaskId         string
}

type AssetDocument struct {
	Id bson.ObjectID `bson:"_id"`

	// 基础属性（默认是不变更的）
	Authority string `bson:"authority"`        //根据URI定义authority = [userinfo@]host[:port]； userinfo本项目里不使用； host包含ip和域名两种方式
	Host      string `bson:"host"`             // host包含ip和域名两种方式
	Port      int    `bson:"port"`             // 端口号，ip是必须的，domain可以为空
	Category  string `bson:"category"`         // host类型：ipv4/ipv6/domain（通过type减少后端的判断）
	Ip        IP     `bson:"ip"`               // host的ip信息
	Domain    string `bson:"domain,omitempty"` // domain是主域名，不包含子域名；如host是sh.doc.baidu.com，则domain是baidu.com

	// 重要的属性：
	Service string   `bson:"service,omitempty"`
	Server  string   `bson:"server,omitempty"`
	Banner  string   `bson:"banner,omitempty"`
	Title   string   `bson:"title,omitempty"`
	App     []string `bson:"app,omitempty"`

	// HTTP相关属性
	HttpStatus string `bson:"status,omitempty"`
	HttpHeader string `bson:"header,omitempty"`
	HttpBody   string `bson:"body,omitempty"`

	Cert     string `bson:"cert,omitempty"`      // 证书信息
	IconHash string `bson:"icon_hash,omitempty"` // 图标hash值
	// 图标hash值文件名和字节内容，用于存储到mongodb，避免每次都从图标文件中读取
	IconHashFile  string `bson:"icon_hash_file,omitempty"`
	IconHashBytes []byte `bson:"icon_hash_bytes,omitempty"`
	OrgId         string `bson:"org,omitempty"`   // 资产所属组织
	ColorTag      string `bson:"color,omitempty"` //颜色标记
	Memo          string `bson:"memo,omitempty"`  //备注
	IsCDN         bool   `bson:"cdn,omitempty"`
	CName         string `bson:"cname,omitempty"`

	IsNewAsset bool   `bson:"new"`    // 是否是新资产
	IsUpdated  bool   `bson:"update"` // 是否有更新
	TaskId     string `bson:"taskId"` // 任务ID

	CreateTime time.Time `bson:"create_time"`
	UpdateTime time.Time `bson:"update_time"`
}

type IP struct {
	IpV4 []IPV4 `bson:"ipv4,omitempty" json:"ipv4,omitempty"`
	IpV6 []IPV6 `bson:"ipv6,omitempty" json:"ipv6,omitempty"`
}

type IPV4 struct {
	IPName   string `bson:"ip" json:"ip"`        //ip的点分十六制字符串表示
	IPInt    uint32 `bson:"uint32" json:"uin32"` //ip的整形表示，为了实现通过子网掩码的检索
	Location string `bson:"location" json:"location"`
}

type IPV6 struct {
	IPName    string `bson:"ip" json:"ip"`                 //ipv6的十六制字符串表示
	IPIntHigh uint64 `bson:"uint64high" json:"uint64high"` //128位整形的高64位
	IPIntLow  uint64 `bson:"uint64low" json:"uint64low"`   //128位整形的低64位；
	Location  string `bson:"location" json:"location"`
}
type StatisticData struct {
	Field string `bson:"_id"`
	Count int    `bson:"count"`
}

type DataSaveStatus struct {
	IsSuccess bool
	IsNew     bool
	IsUpdated bool
}

type DailyCount struct {
	Day   string `bson:"day" json:"day"`
	Count int    `bson:"count" json:"count"`
}

type DailyStat struct {
	ID struct {
		Day string `bson:"day"`
	} `bson:"_id"`
	Count int `bson:"count"`
}

type CategoryStats struct {
	Category    string       `bson:"category" json:"category"`
	DailyCounts []DailyCount `bson:"dailyCounts" json:"dailyCounts"`
	Total       int          `bson:"total" json:"total"`
}

type ApexChartSeries struct {
	Name string `json:"name"`
	Data []int  `json:"data"`
}

type ApexChartData struct {
	Categories []string          `json:"categories"` // 这里存储的是"日"部分，如 ["10", "11", "12"]
	FullDates  []string          `json:"-"`          // 内部使用的完整日期，不暴露给前端
	Series     []ApexChartSeries `json:"series"`
}

func DiffAsset(oldDoc *AssetDocument, newDoc *AssetDocument) (update bson.M, updateCount int) {
	update = make(bson.M)

	if newDoc.HttpStatus != "" && oldDoc.HttpStatus != newDoc.HttpStatus {
		update["status"] = newDoc.HttpStatus
		updateCount++
	}
	// 由于header和body经常变化，所以不统计为更新差异
	if newDoc.HttpHeader != "" && oldDoc.HttpHeader != newDoc.HttpHeader {
		update["header"] = newDoc.HttpHeader
	}
	if newDoc.HttpBody != "" && oldDoc.HttpBody != newDoc.HttpBody {
		update["body"] = newDoc.HttpBody
	}
	if newDoc.Cert != "" && oldDoc.Cert != newDoc.Cert {
		update["cert"] = newDoc.Cert
		updateCount++
	}
	if newDoc.IconHash != "" && oldDoc.IconHash != newDoc.IconHash {
		update["icon_hash"] = newDoc.IconHash
		updateCount++
	}
	if len(newDoc.IconHashBytes) > 0 && len(newDoc.IconHashBytes) != len(oldDoc.IconHashBytes) {
		update["icon_hash_bytes"] = newDoc.IconHashBytes
		updateCount++
	}
	if newDoc.IconHashFile != "" && oldDoc.IconHashFile != newDoc.IconHashFile {
		update["icon_hash_file"] = newDoc.IconHashFile
		updateCount++
	}
	if newDoc.ColorTag != "" && oldDoc.ColorTag != newDoc.ColorTag {
		update["color"] = newDoc.ColorTag
	}
	if newDoc.Memo != "" && oldDoc.Memo != newDoc.Memo {
		update["memo"] = newDoc.Memo
	}
	if newDoc.OrgId != "" && oldDoc.OrgId != newDoc.OrgId {
		update["org"] = newDoc.OrgId
	}
	if newDoc.Service != "" && oldDoc.Service != newDoc.Service {
		update["service"] = newDoc.Service
		updateCount++
	}
	if newDoc.Server != "" && oldDoc.Server != newDoc.Server {
		update["server"] = newDoc.Server
		updateCount++
	}
	if newDoc.Banner != "" && oldDoc.Banner != newDoc.Banner {
		update["banner"] = newDoc.Banner
		updateCount++
	}
	if newDoc.Title != "" && oldDoc.Title != newDoc.Title {
		update["title"] = newDoc.Title
		updateCount++
	}
	if len(newDoc.App) > 0 && utils.CompareStringSlices(oldDoc.App, newDoc.App, true) == false {
		update["app"] = newDoc.App
		updateCount++
	}
	if len(newDoc.CName) > 0 && oldDoc.CName != newDoc.CName {
		update["cname"] = newDoc.CName
		updateCount++
	}
	if newDoc.IsCDN != oldDoc.IsCDN {
		update["cdn"] = newDoc.IsCDN
		updateCount++
	}
	// 更新时间
	update[UpdateTime] = time.Now()

	return
}

// NewAsset 创建一个新的实例
func NewAsset(databaseName, collectionName string, taskId string, client *mongo.Client) *Asset {
	return &Asset{
		DatabaseName:   databaseName,
		CollectionName: collectionName,
		TaskId:         taskId,
		Ctx:            context.Background(),
		Client:         client,
	}
}

func (fa *Asset) Insert(doc AssetDocument) (isSuccess bool, err error) {
	// 生成_id
	if doc.Id.IsZero() {
		doc.Id = bson.NewObjectID()
	}
	// 处理ipv6与ipv6
	if len(doc.Ip.IpV4) > 0 {
		for i := 0; i < len(doc.Ip.IpV4); i++ {
			doc.Ip.IpV4[i].IPInt = utils.IPV4ToUInt32(doc.Ip.IpV4[i].IPName)
		}
	}
	if len(doc.Ip.IpV6) > 0 {
		for i := 0; i < len(doc.Ip.IpV6); i++ {
			doc.Ip.IpV6[i].IPIntHigh, doc.Ip.IpV6[i].IPIntLow = utils.IPV6ToDoubleInt64(doc.Ip.IpV6[i].IPName)
		}
	}
	// 时间
	now := time.Now()
	doc.IsNewAsset = true
	doc.CreateTime = now
	doc.UpdateTime = now
	// 插入文档
	col := fa.Client.Database(fa.DatabaseName).Collection(fa.CollectionName)
	_, err = col.InsertOne(fa.Ctx, doc)
	if err != nil {
		return
	}
	isSuccess = true
	return
}

func (fa *Asset) Delete(id string) (isSuccess bool, err error) {
	_idd, _ := bson.ObjectIDFromHex(id)
	col := fa.Client.Database(fa.DatabaseName).Collection(fa.CollectionName)
	_, err = col.DeleteOne(fa.Ctx, bson.M{ID: _idd})
	if err != nil {
		return false, err
	}
	isSuccess = true
	return
}

func (fa *Asset) DeleteByTaskId(taskId string) (isSuccess bool, err error) {
	col := fa.Client.Database(fa.DatabaseName).Collection(fa.CollectionName)
	_, err = col.DeleteMany(fa.Ctx, bson.M{TaskId: taskId})
	if err != nil {
		return false, err
	}
	isSuccess = true
	return
}

func (fa *Asset) DeleteByHost(host string) (isSuccess bool, err error) {
	col := fa.Client.Database(fa.DatabaseName).Collection(fa.CollectionName)
	_, err = col.DeleteMany(fa.Ctx, bson.M{"host": host})
	if err != nil {
		return false, err
	}
	isSuccess = true
	return
}

func (fa *Asset) FindByAuthority(authority string) (doc *AssetDocument, err error) {
	doc = &AssetDocument{}
	col := fa.Client.Database(fa.DatabaseName).Collection(fa.CollectionName)
	filter := bson.D{{"authority", authority}, {TaskId, fa.TaskId}}
	err = col.FindOne(fa.Ctx, filter).Decode(doc)
	if err != nil {
		return nil, err
	}
	return
}

func (fa *Asset) Update(id string, update bson.M) (isSuccess bool, err error) {
	_idd, _ := bson.ObjectIDFromHex(id)
	col := fa.Client.Database(fa.DatabaseName).Collection(fa.CollectionName)
	filter := bson.D{{ID, _idd}}
	updateDoc := bson.D{{"$set", update}}
	_, err = col.UpdateOne(fa.Ctx, filter, updateDoc)
	if err != nil {
		return
	}
	isSuccess = true
	return
}

func (fa *Asset) InsertOrUpdate(doc AssetDocument) (dss DataSaveStatus, err error) {
	oldDoc, err := fa.FindByAuthority(doc.Authority)
	if oldDoc == nil {
		dss.IsNew = true
		dss.IsSuccess, err = fa.Insert(doc)
		return
	}
	// 更新文档
	update, updateCount := DiffAsset(oldDoc, &doc)
	update["new"] = false
	if updateCount > 0 {
		update["update"] = true
		dss.IsUpdated = true
	}
	filter := bson.D{{ID, oldDoc.Id}}
	updateDoc := bson.D{{"$set", update}}
	col := fa.Client.Database(fa.DatabaseName).Collection(fa.CollectionName)
	_, err = col.UpdateOne(fa.Ctx, filter, updateDoc)
	if err != nil {
		return
	}
	dss.IsSuccess = true

	return
}

func (fa *Asset) Find(filter bson.M, page, rowsPerPage int, sortByDate bool, excludeHttpBody bool) (docs []AssetDocument, err error) {
	opts := options.Find()
	if page > 0 && rowsPerPage > 0 {
		opts.SetSkip(int64((page - 1) * rowsPerPage))
		opts.SetLimit(int64(rowsPerPage))
	}
	if sortByDate {
		opts.SetSort(bson.D{{Key: UpdateTime, Value: -1}, {Key: "authority", Value: 1}})
	} else {
		opts.SetSort(bson.D{{Key: "authority", Value: 1}})
	}
	if excludeHttpBody {
		opts.SetProjection(bson.D{{Key: "body", Value: 0}})
	}
	col := fa.Client.Database(fa.DatabaseName).Collection(fa.CollectionName)
	cursor, err := col.Find(fa.Ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer func(cursor *mongo.Cursor, ctx context.Context) {
		_ = cursor.Close(ctx)
	}(cursor, fa.Ctx)

	if err = cursor.All(fa.Ctx, &docs); err != nil {
		return nil, err
	}
	return
}

func (fa *Asset) Count(filter bson.M) (int, error) {
	col := fa.Client.Database(fa.DatabaseName).Collection(fa.CollectionName)
	count, err := col.CountDocuments(fa.Ctx, filter)
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

func (fa *Asset) Get(id string) (doc AssetDocument, err error) {
	col := fa.Client.Database(fa.DatabaseName).Collection(fa.CollectionName)
	_idd, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return
	}
	filter := bson.D{{ID, _idd}}
	err = col.FindOne(fa.Ctx, filter).Decode(&doc)
	return
}

func (fa *Asset) Aggregate(filter bson.M, field string, limit int, unwind bool) (results []StatisticData, err error) {
	// 获取集合
	collection := fa.Client.Database(fa.DatabaseName).Collection(fa.CollectionName)
	// 定义聚合函数
	var pipeline mongo.Pipeline

	// 如果需要展开数组
	if unwind {
		pipeline = append(pipeline, bson.D{{"$unwind", "$" + field}})
	}
	pipeline = append(pipeline, bson.D{{"$match", filter}})
	// 对 port 字段特殊处理，转换为字符串
	if field == "port" {
		pipeline = append(pipeline, bson.D{
			{"$group", bson.D{
				{"_id", bson.D{{"$toString", "$" + field}}}, // 将 port 转换为字符串
				{"count", bson.D{{"$sum", 1}}},
			}}},
		)
	} else {
		pipeline = append(pipeline, bson.D{
			{"$group", bson.D{
				{"_id", "$" + field},
				{"count", bson.D{{"$sum", 1}}},
			}}},
		)
	}
	// 排序和限制
	pipeline = append(pipeline, bson.D{{"$sort", bson.D{{"count", -1}}}})
	if limit > 0 {
		pipeline = append(pipeline, bson.D{{"$limit", limit}})
	}
	// 执行聚合查询
	cursor, err := collection.Aggregate(fa.Ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer func(cursor *mongo.Cursor, ctx context.Context) {
		_ = cursor.Close(ctx)
	}(cursor, fa.Ctx)

	// 解析结果为 StatisticData 数组
	err = cursor.All(fa.Ctx, &results)
	return
}

// GenerateChartData 生成图表数据，days参数指定要统计的天数
func (fa *Asset) GenerateChartData(days int) (ApexChartData, error) {
	// 1. 获取原始统计数据
	stats, err := fa.getCategoryStats(days)
	if err != nil {
		return ApexChartData{}, err
	}

	// 2. 转换为ApexCharts数据
	return convertToApexChartData(stats, days)
}

func (fa *Asset) getCategoryStats(days int) ([]CategoryStats, error) {
	collection := fa.Client.Database(fa.DatabaseName).Collection(fa.CollectionName)

	startDate := time.Now().Add(-time.Duration(days) * 24 * time.Hour)
	pipeline := mongo.Pipeline{
		{{"$match", bson.D{
			{"create_time", bson.D{
				{"$gte", startDate},
				{"$lte", time.Now()},
			}},
		}}},
		{{"$project", bson.D{
			{"category", 1},
			{"day", bson.D{
				{"$dateToString", bson.D{
					{"format", "%Y-%m-%d"},
					{"date", "$create_time"},
				}},
			}},
		}}},
		{{"$group", bson.D{
			{"_id", bson.D{
				{"category", "$category"},
				{"day", "$day"},
			}},
			{"count", bson.D{{"$sum", 1}}},
		}}},
		{{"$sort", bson.D{
			{"_id.category", 1},
			{"_id.day", 1},
		}}},
		{{"$group", bson.D{
			{"_id", "$_id.category"},
			{"dailyCounts", bson.D{
				{"$push", bson.D{
					{"day", "$_id.day"},
					{"count", "$count"},
				}},
			}},
			{"total", bson.D{{"$sum", "$count"}}},
		}}},
		{{"$project", bson.D{
			{"category", "$_id"},
			{"dailyCounts", 1},
			{"total", 1},
			{"_id", 0},
		}}},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []CategoryStats
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	return results, nil
}

func convertToApexChartData(stats []CategoryStats, days int) (ApexChartData, error) {
	// 生成完整的日期范围(从days天前到今天)
	now := time.Now()
	startDate := now.Add(-time.Duration(days) * 24 * time.Hour)

	// 准备两个切片：
	// fullDates 存储完整日期 ["2025-03-10", "2025-03-11", ...]
	// dayOnly 存储只有日的部分 ["10", "11", ...]
	fullDates := make([]string, 0, days)
	dayOnly := make([]string, 0, days)

	for d := startDate; !d.After(now); d = d.Add(24 * time.Hour) {
		fullDate := d.Format("2006-01-02")
		fullDates = append(fullDates, fullDate)
		dayOnly = append(dayOnly, d.Format("02")) // 只取"日"部分
	}

	// 为每个category创建series
	series := make([]ApexChartSeries, 0, len(stats))
	for _, category := range stats {
		// 创建日期到计数的映射
		countMap := make(map[string]int)
		for _, daily := range category.DailyCounts {
			countMap[daily.Day] = daily.Count
		}

		// 按完整日期范围填充数据
		data := make([]int, 0, len(fullDates))
		for _, date := range fullDates {
			if count, exists := countMap[date]; exists {
				data = append(data, count)
			} else {
				data = append(data, 0) // 没有数据的日期填充0
			}
		}

		series = append(series, ApexChartSeries{
			Name: category.Category,
			Data: data,
		})
	}

	return ApexChartData{
		Categories: dayOnly,   // 前端显示只用"日"部分
		FullDates:  fullDates, // 保留完整日期供可能的后端处理使用
		Series:     series,
	}, nil
}

// ExportToCSV 查询MongoDB并将结果导出为CSV文件
func (fa *Asset) ExportToCSV(filter bson.M, outputFile string, fields []string) (int64, error) {
	collection := fa.Client.Database(fa.DatabaseName).Collection(fa.CollectionName)
	// 2. 执行查询
	opts := options.Find()
	opts.SetProjection(bson.D{{Key: "body", Value: 0}})
	cur, err := collection.Find(fa.Ctx, filter, opts)
	if err != nil {
		return 0, err
	}
	defer cur.Close(fa.Ctx)

	// 3. 创建CSV文件
	file, err := os.Create(outputFile)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// 4. 处理字段选择
	allFields := getAllFieldNames()
	selectedFields := allFields
	if len(fields) > 0 {
		selectedFields = fields
	}

	// 5. 写入CSV头
	if err := writer.Write(selectedFields); err != nil {
		return 0, err
	}

	// 6. 遍历结果并写入CSV
	var count int64
	for cur.Next(fa.Ctx) {
		var doc AssetDocument
		if err := cur.Decode(&doc); err != nil {
			return count, err
		}

		record, err := docToRecord(&doc, selectedFields)
		if err != nil {
			return count, err
		}

		if err := writer.Write(record); err != nil {
			return count, err
		}

		count++
	}

	if err := cur.Err(); err != nil {
		return count, err
	}

	return count, nil
}

// getAllFieldNames 返回所有可能的字段名
func getAllFieldNames() []string {
	return []string{
		"authority", "host", "port", "category",
		"ip.ipv4.ip", "ip.ipv4.location", "ip.ipv6.ip", "ip.ipv6.location",
		"domain", "service", "server", "banner", "title", "app",
		"status", "header", "cert", "icon_hash",
		"color", "memo", "cdn", "cname",
		"create_time", "update_time",
	}
}

// docToRecord 将文档转换为CSV记录
func docToRecord(doc *AssetDocument, fields []string) ([]string, error) {
	record := make([]string, len(fields))

	for i, field := range fields {
		switch field {
		case "authority":
			record[i] = doc.Authority
		case "host":
			record[i] = doc.Host
		case "port":
			record[i] = fmt.Sprintf("%d", doc.Port)
		case "category":
			record[i] = doc.Category
		case "ip.ipv4.ip":
			if len(doc.Ip.IpV4) > 0 {
				var ipArray []string
				for _, ip := range doc.Ip.IpV4 {
					ipArray = append(ipArray, ip.IPName)
				}
				record[i] = strings.Join(ipArray, ",")
			}
		case "ip.ipv4.location":
			if len(doc.Ip.IpV4) > 0 {
				var ipArray []string
				for _, ip := range doc.Ip.IpV4 {
					ipArray = append(ipArray, ip.Location)
				}
				record[i] = strings.Join(ipArray, ",")
			}
		case "ip.ipv6.ip":
			if len(doc.Ip.IpV6) > 0 {
				var ipArray []string
				for _, ip := range doc.Ip.IpV6 {
					ipArray = append(ipArray, ip.IPName)
				}
				record[i] = strings.Join(ipArray, ",")
			}
		case "ip.ipv6.location":
			if len(doc.Ip.IpV6) > 0 {
				var ipArray []string
				for _, ip := range doc.Ip.IpV6 {
					ipArray = append(ipArray, ip.Location)
				}
				record[i] = strings.Join(ipArray, ",")
			}
		case "domain":
			record[i] = doc.Domain
		case "service":
			record[i] = doc.Service
		case "server":
			record[i] = doc.Server
		case "banner":
			record[i] = doc.Banner
		case "title":
			record[i] = doc.Title
		case "app":
			record[i] = strings.Join(doc.App, ",")
		case "status":
			record[i] = doc.HttpStatus
		case "header":
			record[i] = doc.HttpHeader
		case "cert":
			record[i] = doc.Cert
		case "icon_hash":
			record[i] = doc.IconHash
		case "color":
			record[i] = doc.ColorTag
		case "memo":
			record[i] = doc.Memo
		case "cdn":
			record[i] = strconv.FormatBool(doc.IsCDN)
		case "cname":
			record[i] = doc.CName
		case "create_time":
			record[i] = doc.CreateTime.In(conf.LocalTimeLocation).Format(time.RFC3339)
		case "update_time":
			record[i] = doc.UpdateTime.In(conf.LocalTimeLocation).Format(time.RFC3339)
		default:
			return nil, fmt.Errorf("unknown field: %s", field)
		}
	}

	return record, nil
}

// ImportFromCSV 将CSV文件导入到MongoDB
func (fa *Asset) ImportFromCSV(filename string, orgId string, funcCall RecordToDocFunc, insertAndUpdate bool) (int64, error) {
	// 1. 打开CSV文件
	file, err := os.Open(filename)
	if err != nil {
		return 0, fmt.Errorf("failed to open CSV file: %v", err)
	}
	defer file.Close()

	// 2. 创建CSV reader
	reader := csv.NewReader(file)

	// 3. 读取表头
	headers, err := reader.Read()
	if err != nil {
		return 0, fmt.Errorf("failed to read CSV headers: %v", err)
	}
	// 4. 准备批量写入
	collection := fa.Client.Database(fa.DatabaseName).Collection(fa.CollectionName)
	var documents []AssetDocument
	var totalInserted int64
	batchSize := 100 // 每批插入100条记录

	// 5. 逐行读取并转换
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return totalInserted, fmt.Errorf("CSV read error: %v", err)
		}

		// 6. 将CSV记录转换为AssetDocument
		var doc *AssetDocument
		doc, err = funcCall(headers, record, orgId)
		//if err != nil {
		//	fmt.Println(err)
		//}
		if doc == nil {
			continue // 跳过空记录
		}
		documents = append(documents, *doc)

		// 7. 批量插入
		if len(documents) >= batchSize {
			if insertAndUpdate {
				count, err := fa.importByUpdate(documents)
				if err != nil {
					return totalInserted, err
				}
				totalInserted += count
				totalInserted += count
			} else {
				result, err := collection.InsertMany(fa.Ctx, documents)
				if err != nil {
					return totalInserted, fmt.Errorf("batch insert failed: %v", err)
				}
				totalInserted += int64(len(result.InsertedIDs))
				documents = documents[:0] // 清空切片
			}
		}
	}

	// 8. 插入剩余记录
	if len(documents) > 0 {
		if insertAndUpdate {
			count, err := fa.importByUpdate(documents)
			if err != nil {
				return totalInserted, err
			}
			totalInserted += count
		} else {
			result, err := collection.InsertMany(fa.Ctx, documents)
			if err != nil {
				return totalInserted, fmt.Errorf("final batch insert failed: %v", err)
			}
			totalInserted += int64(len(result.InsertedIDs))
		}
	}

	return totalInserted, nil
}

func (fa *Asset) importByUpdate(docs []AssetDocument) (count int64, err error) {
	for _, doc := range docs {
		dss, err := fa.InsertOrUpdate(doc)
		if err != nil {
			return 0, err
		}
		if dss.IsSuccess {
			count++
		}
	}
	return count, nil
}

// RecordToDoc 将CSV记录转换为AssetDocument
func RecordToDoc(headers, record []string, orgId string) (*AssetDocument, error) {
	now := time.Now()
	doc := &AssetDocument{
		Id:         bson.NewObjectID(),
		OrgId:      orgId,
		CreateTime: now,
		UpdateTime: now,
	}

	for i, header := range headers {
		if i >= len(record) {
			continue // 跳过不完整的记录
		}

		value := record[i]
		if value == "" {
			continue // 跳过空值
		}

		switch header {
		case "authority":
			doc.Authority = value
		case "host":
			doc.Host = value
		case "port":
			if port, err := strconv.Atoi(value); err == nil {
				doc.Port = port
			}
		case "category":
			doc.Category = value
		case "ip.ipv4.ip":
			ips := strings.Split(value, ",")
			for _, ip := range ips {
				if ip != "" {
					doc.Ip.IpV4 = append(doc.Ip.IpV4, IPV4{IPName: ip, IPInt: utils.IPV4ToUInt32(ip)})
				}
			}
		case "ip.ipv4.location":
			locations := strings.Split(value, ",")
			for i, loc := range locations {
				if i < len(doc.Ip.IpV4) && loc != "" {
					doc.Ip.IpV4[i].Location = loc
				}
			}
		case "ip.ipv6.ip":
			ips := strings.Split(value, ",")
			for _, ip := range ips {
				if ip != "" {
					h, l := utils.IPV6ToDoubleInt64(ip)
					doc.Ip.IpV6 = append(doc.Ip.IpV6, IPV6{IPName: ip, IPIntHigh: h, IPIntLow: l})
				}
			}
		case "ip.ipv6.location":
			locations := strings.Split(value, ",")
			for i, loc := range locations {
				if i < len(doc.Ip.IpV6) && loc != "" {
					doc.Ip.IpV6[i].Location = loc
				}
			}
		case "domain":
			doc.Domain = value
		case "service":
			doc.Service = value
		case "server":
			doc.Server = value
		case "banner":
			doc.Banner = value
		case "title":
			doc.Title = value
		case "app":
			doc.App = strings.Split(value, ",")
		case "status":
			doc.HttpStatus = value
		case "header":
			doc.HttpHeader = value
		case "cert":
			doc.Cert = value
		case "icon_hash":
			doc.IconHash = value
		case "color":
			doc.ColorTag = value
		case "memo":
			doc.Memo = value
		case "cdn":
			if isCDN, err := strconv.ParseBool(value); err == nil {
				doc.IsCDN = isCDN
			}
		case "cname":
			doc.CName = value
		case "create_time":
			if t, err := time.Parse(time.RFC3339, value); err == nil {
				doc.CreateTime = t
			}
		case "update_time":
			if t, err := time.Parse(time.RFC3339, value); err == nil {
				doc.UpdateTime = t
			}
		}
	}
	if len(doc.Authority) == 0 || len(doc.Host) == 0 || len(doc.Category) == 0 {
		return nil, errors.New("authority, host, category is empty")
	}
	if len(doc.Ip.IpV4) == 0 && len(doc.Ip.IpV6) == 0 {
		return nil, errors.New("ip is empty")
	}

	return doc, nil
}
