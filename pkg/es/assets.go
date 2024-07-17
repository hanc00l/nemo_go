package es

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"github.com/cenkalti/backoff/v4"
	"github.com/dustin/go-humanize"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esutil"
	"github.com/elastic/go-elasticsearch/v8/typedapi/core/search"
	"github.com/elastic/go-elasticsearch/v8/typedapi/indices/create"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types/enums/sortorder"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"runtime"
	"strings"
	"sync/atomic"
	"time"
)

var ShowBulkIndexStatistic bool

type Assets struct {
	IndexName string
	Ctx       context.Context
}

type Document struct {
	Id         string    `json:"id"`
	Host       string    `json:"host"`
	Ip         []string  `json:"ip"`
	Port       int       `json:"port"`
	Domain     string    `json:"domain"`
	Location   []string  `json:"location"`
	Status     int       `json:"status"`
	Service    string    `json:"service"`
	Banner     string    `json:"banner"`
	Server     string    `json:"server"`
	Title      string    `json:"title"`
	Header     string    `json:"header"`
	Body       string    `json:"body"`
	Cert       string    `json:"cert"`
	IconHash   int64     `json:"icon_hash"`
	Org        string    `json:"org"`
	Source     []string  `json:"source"`
	Comment    string    `json:"comment"`
	CreateTime time.Time `json:"create_time"`
	UpdateTime time.Time `json:"update_time"`
}

// NewAssets 创建对象
func NewAssets(indexName string) *Assets {
	return &Assets{
		IndexName: indexName,
		Ctx:       context.Background(),
	}
}

// GetIndexMapping 定义索引的mapping
func (a *Assets) GetIndexMapping() *types.TypeMapping {
	//指定使用分词器
	analyzer := "stop"
	hostProperty := types.NewTextProperty()
	hostProperty.Analyzer = &analyzer
	hostProperty.SearchAnalyzer = &analyzer

	return &types.TypeMapping{
		Properties: map[string]types.Property{
			"id":     types.NewKeywordProperty(),       // sha1(host)
			"host":   hostProperty,                     // host 保存子域名（主机名）加端口；对于域名资资产就是www.163.com，对ip就是114.114.114.114:53；host具有唯一性
			"ip":     types.NewIpProperty(),            // ip ipv4/6，对于域名资产是域名解析的ip（可关联多个，因此是数组），对于ip资产就是ip
			"port":   types.NewIntegerNumberProperty(), // port 端口号
			"domain": types.NewKeywordProperty(),       // domain keyword类型，只保存主域名，用于查询结果的分类的排序用
			// 属性
			"location":  types.NewTextProperty(),
			"status":    types.NewIntegerNumberProperty(),
			"service":   types.NewTextProperty(),
			"banner":    types.NewTextProperty(),
			"server":    types.NewTextProperty(),
			"title":     types.NewTextProperty(),
			"header":    types.NewTextProperty(),
			"body":      types.NewTextProperty(),
			"cert":      types.NewTextProperty(),
			"icon_hash": types.NewLongNumberProperty(),
			"org":       types.NewTextProperty(), // 关联的组织名称
			"source":    types.NewKeywordProperty(),
			"comment":   types.NewTextProperty(),
			// 时间记录
			"create_time": types.NewDateProperty(),
			"update_time": types.NewDateProperty(),
		},
	}
}

func (a *Assets) GetIndexMappingWithJson() string {
	mappingJSON := `{
        "mappings": {
            "properties": {
                "banner": {
                    "type": "text"
                },
                "body": {
                    "type": "text"
                },
                "cert": {
                    "type": "text"
                },
                "comment": {
                    "type": "text"
                },
               
                "header": {
                    "type": "text"
                },
                "host": {
                    "type": "text",
                    "analyzer": "stop"
                },
                "icon_hash": {
                    "type": "long"
                },
                "id": {
                    "type": "keyword"
                },
                "ip": {
                    "type": "ip"
                },
                "location": {
                    "type": "text"
                },
                "org": {
                    "type": "text"
                },
                "port": {
                    "type": "integer"
                },
                "server": {
                    "type": "text"
                },
                "service": {
                    "type": "text"
                },
                "source": {
                    "type": "keyword"
                },
                "status": {
                    "type": "integer"
                },
                "title": {
                    "type": "text"
                },
                "update_time": {
                    "type": "date"
                }
            }
        }
	}`
	return mappingJSON
}

// CreateIndex 创建索引
func (a *Assets) CreateIndex() bool {
	// 检查索引是否存在
	exists, err := GetTypedClient().Indices.Exists(a.IndexName).IsSuccess(a.Ctx)
	if err != nil {
		logging.RuntimeLog.Errorf("check index exist failed:%v", err)
		return false
	}
	if !exists {
		_, err = GetTypedClient().
			Indices.
			Create(a.IndexName).
			Request(&create.Request{
				Mappings: a.GetIndexMapping(),
			}).Do(a.Ctx)
		if err != nil {
			logging.RuntimeLog.Errorf("create index failed:%v", err)
			return false
		}
	}
	return true
}

func (a *Assets) CreateIndexWithJsonMapping() bool {
	client, err := elasticsearch.NewClient(GetElasticConfig())
	if err != nil {
		logging.RuntimeLog.Errorf("Error creating the client: %s", err)
		return false
	}
	// 检查索引是否存在
	resp, err := client.Indices.Exists([]string{a.IndexName},
		client.Indices.Exists.WithContext(a.Ctx))
	if err != nil {
		logging.RuntimeLog.Errorf("Error checking the index: %s", err)
		return false
	}
	if !resp.IsError() {
		logging.RuntimeLog.Info("Index already exists")
		return true
	}
	// 创建索引，使用json格式的mapping
	_, err = client.Indices.Create(a.IndexName,
		client.Indices.Create.WithContext(a.Ctx),
		client.Indices.Create.WithBody(strings.NewReader(a.GetIndexMappingWithJson())),
	)
	if err != nil {
		logging.RuntimeLog.Errorf("Error creating the index: %s", err)
		return false
	}
	return true
}

// DeleteIndex 删除索引
func (a *Assets) DeleteIndex() bool {
	_, err := GetTypedClient().Indices.
		Delete(a.IndexName).
		Do(a.Ctx)
	if err != nil {
		logging.RuntimeLog.Errorf("delete index failed:%v", err)
		return false
	}
	return true
}

// IndexDoc 索引文档
func (a *Assets) IndexDoc(doc Document) bool {
	doc.Id = SID(doc.Host)
	_, err := GetTypedClient().
		Index(a.IndexName).
		Id(doc.Id).
		Request(doc).
		Do(a.Ctx)
	if err != nil {
		logging.RuntimeLog.Errorf("index doc failed:%v", err)
		return false
	}
	return true
}

// CheckDoc 检查文档是否存在
func (a *Assets) CheckDoc(docId string) bool {
	exists, err := GetTypedClient().
		Exists(a.IndexName, docId).
		Do(a.Ctx)
	if err != nil {
		logging.RuntimeLog.Errorf("check index exist failed:%v", err)
		return false
	}
	return exists
}

// GetDoc 根据id获取一个文档对象
func (a *Assets) GetDoc(docId string) (doc Document, status bool) {
	res, err := GetTypedClient().
		Get(a.IndexName, docId).
		Do(a.Ctx)
	if err != nil {
		logging.RuntimeLog.Errorf("get doc failed:%v", err)
		return
	}
	if !res.Found {
		return
	}
	err = json.Unmarshal(res.Source_, &doc)
	if err != nil {
		logging.RuntimeLog.Errorf("unmarshal doc failed:%v", err)
		return doc, false
	}
	return doc, true
}

// DeleteDoc 根据id删除一个文档对象
func (a *Assets) DeleteDoc(docId string) (status bool) {
	_, err := GetTypedClient().
		Delete(a.IndexName, docId).
		Do(a.Ctx)
	if err != nil {
		logging.RuntimeLog.Errorf("delete doc failed:%v", err)
		return
	}
	return true
}

// Search 查询，根据查询条件返回指定数量的文档，查询使用是bool查询
func (a *Assets) Search(query types.Query, page, rowsPerPage int) (res *search.Response, err error) {
	req := &search.Request{
		Query: &query,
	}
	if rowsPerPage > 0 && page > 0 {
		pageFrom := rowsPerPage * (page - 1)
		req.From = &pageFrom
		req.Size = &rowsPerPage

	}
	req.Sort = []types.SortCombinations{
		types.SortOptions{SortOptions: map[string]types.FieldSort{
			"update_time": {Order: &sortorder.Desc},
		}},
	}
	req.TrackTotalHits = types.TrackHits(true)

	res, err = GetTypedClient().Search().
		Index(a.IndexName).
		Request(req).
		Do(a.Ctx)
	if err != nil {
		logging.RuntimeLog.Errorf("search failed:%v", err)
		return
	}

	return
}

// BulkIndexDoc 批量索文档
func (a *Assets) BulkIndexDoc(docs []Document) bool {
	var countSuccessful uint64
	cfg := GetElasticConfig()
	retryBackoff := backoff.NewExponentialBackOff()
	cfg.RetryBackoff = func(i int) time.Duration {
		if i == 1 {
			retryBackoff.Reset()
		}
		return retryBackoff.NextBackOff()
	}
	cfg.RetryOnStatus = []int{502, 503, 504, 429}
	cfg.MaxRetries = 5
	// bulk不支持typedClient
	client, err := elasticsearch.NewClient(cfg)
	if err != nil {
		logging.RuntimeLog.Errorf("Error creating the client: %s", err)
		return false
	}
	bi, err := esutil.NewBulkIndexer(esutil.BulkIndexerConfig{
		Index:         a.IndexName,      // The default index name
		Client:        client,           // The Elasticsearch client
		NumWorkers:    runtime.NumCPU(), // The number of worker goroutines
		FlushBytes:    5e+6,             // The flush threshold in bytes
		FlushInterval: 30 * time.Second, // The periodic flush interval
	})
	if err != nil {
		logging.RuntimeLog.Errorf("Error creating the indexer: %s", err)
		return false
	}
	start := time.Now().UTC()
	// Loop over the collection
	for _, doc := range docs {
		// Prepare the data payload: encode article to JSON
		data, err := json.Marshal(doc)
		if err != nil {
			logging.RuntimeLog.Errorf("Cannot encode article %s: %s", doc.Id, err)
			continue
		}
		// Add the document to the indexer
		err = bi.Add(
			context.Background(),
			esutil.BulkIndexerItem{
				// Action field configures the operation to perform (index, create, delete, update)
				Action: "index",
				// DocumentID is the (optional) document ID
				DocumentID: doc.Id,
				// Body is an `io.Reader` with the payload
				Body: bytes.NewReader(data),
				// OnSuccess is called for each successful operation
				OnSuccess: func(ctx context.Context, item esutil.BulkIndexerItem, res esutil.BulkIndexerResponseItem) {
					atomic.AddUint64(&countSuccessful, 1)
				},
				// OnFailure is called for each failed operation
				OnFailure: func(ctx context.Context, item esutil.BulkIndexerItem, res esutil.BulkIndexerResponseItem, err error) {
					if err != nil {
						logging.RuntimeLog.Errorf("ERROR: %s", err)
					} else {
						logging.RuntimeLog.Errorf("ERROR: %s: %s", res.Error.Type, res.Error.Reason)
					}
				},
			},
		)
		if err != nil {
			logging.RuntimeLog.Errorf("Unexpected error: %s", err)
			return false
		}
	}
	if err := bi.Close(context.Background()); err != nil {
		logging.RuntimeLog.Errorf("Unexpected error: %s", err)
		return false
	}
	biStats := bi.Stats()
	if ShowBulkIndexStatistic {
		// Report the results: number of indexed docs, number of errors, duration, indexing rate
		dur := time.Since(start)
		if biStats.NumFailed > 0 {
			logging.CLILog.Errorf(
				"Indexed [%s] documents with [%s] errors in %s (%s docs/sec)",
				humanize.Comma(int64(biStats.NumFlushed)),
				humanize.Comma(int64(biStats.NumFailed)),
				dur.Truncate(time.Millisecond),
				humanize.Comma(int64(1000.0/float64(dur/time.Millisecond)*float64(biStats.NumFlushed))),
			)
		} else {
			logging.CLILog.Infof(
				"Sucessfuly indexed [%s] documents in %s (%s docs/sec)",
				humanize.Comma(int64(biStats.NumFlushed)),
				dur.Truncate(time.Millisecond),
				humanize.Comma(int64(1000.0/float64(dur/time.Millisecond)*float64(biStats.NumFlushed))),
			)
		}
	}
	return true
}

// SID 生成唯一标识
func SID(plainText string) string {
	sha1 := sha1.New()
	sha1.Write([]byte(plainText))
	return hex.EncodeToString(sha1.Sum(nil))
}
