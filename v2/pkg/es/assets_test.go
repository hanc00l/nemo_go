package es

import (
	"encoding/json"
	"github.com/elastic/go-elasticsearch/v8/typedapi/core/search"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/hanc00l/nemo_go/v2/pkg/utils"
	"testing"
	"time"
)

func outputDoc(res *search.Response, t *testing.T) {
	t.Logf("Hits:%d", res.Hits.Total.Value)
	for _, v := range res.Hits.Hits {
		var doc Document
		json.Unmarshal(v.Source_, &doc)
		out, _ := json.MarshalIndent(doc, "", "    ")
		t.Log(string(out))
	}
}

func TestAssets_CreateIndex(t *testing.T) {
	//assets := NewAssets("test1")
	workspaceResult := map[int]string{
		1: "b0c79065-7ff7-32ae-cc18-864ccd8f7717",
	}
	assets := NewAssets(workspaceResult[1])
	t.Log(assets.CreateIndex())
}

func TestAssets_DeleteIndex(t *testing.T) {
	//assets := NewAssets("test1")
	//本地测试第一个
	workspaceResult := map[int]string{
		1: "b0c79065-7ff7-32ae-cc18-864ccd8f7717",
	}
	assets := NewAssets(workspaceResult[1])
	t.Log(assets.DeleteIndex())
}

func TestAssets_IndexDoc(t *testing.T) {
	assets := NewAssets("test1")
	t.Log(assets.DeleteIndex())
	t.Log(assets.CreateIndex())
	doc1 := Document{
		Host:       "192.168.120.1:443",
		Ip:         []string{"192.168.120.1"},
		Port:       443,
		Domain:     "",
		Location:   []string{"局域网"},
		Status:     200,
		Service:    "https",
		Banner:     "nginx",
		Server:     "nginx",
		Title:      "huawei WIFI",
		Header:     "",
		Body:       "",
		Cert:       "",
		IconHash:   0,
		Org:        "测试机构1",
		Source:     []string{"httpx"},
		Comment:    "",
		CreateTime: time.Now(),
		UpdateTime: time.Now(),
	}
	t.Log(assets.IndexDoc(doc1))
	doc2 := Document{
		Host:       "192.168.120.1:80",
		Ip:         []string{"192.168.120.1"},
		Port:       80,
		Domain:     "",
		Location:   []string{"局域网"},
		Status:     302,
		Service:    "https",
		Banner:     "nginx",
		Server:     "nginx",
		Title:      "正在跳转...huawei WIFI",
		Header:     "",
		Body:       "",
		Cert:       "",
		IconHash:   0,
		Org:        "测试机构1",
		Source:     []string{"httpx"},
		Comment:    "",
		CreateTime: time.Now(),
		UpdateTime: time.Now(),
	}
	t.Log(assets.IndexDoc(doc2))
	doc3 := Document{
		Host:       "test.nemo.com.cn",
		Ip:         []string{"192.168.120.1", "192.168.120.2"},
		Port:       0,
		Domain:     "nemo.com.cn",
		Location:   []string{"中国-北京"},
		Status:     200,
		Service:    "https",
		Banner:     "PHP/8.3",
		Server:     "IIS/8.5",
		Title:      "Nemo Login",
		Header:     "",
		Body:       "",
		Cert:       "",
		IconHash:   0,
		Org:        "测试机构2",
		Source:     []string{"fofa"},
		Comment:    "",
		CreateTime: time.Now(),
		UpdateTime: time.Now(),
	}
	t.Log(assets.IndexDoc(doc3))
	doc4 := Document{
		Host:       "minichat.news.com.cn",
		Ip:         []string{"10.0.0.5"},
		Port:       0,
		Domain:     "nemo.com.cn",
		Location:   []string{"中国-上海"},
		Status:     403,
		Service:    "https",
		Banner:     "X-Powered-By: Apache",
		Server:     "Python/3.10",
		Title:      "Minichat Login",
		Header:     "",
		Body:       "",
		Cert:       "",
		IconHash:   0,
		Org:        "测试机构3",
		Source:     []string{"httpx"},
		Comment:    "",
		CreateTime: time.Now(),
		UpdateTime: time.Now(),
	}
	t.Log(assets.IndexDoc(doc4))
}

func TestAssets_Search(t *testing.T) {
	query := types.Query{
		Bool: &types.BoolQuery{},
	}
	q1 := types.Query{
		MatchPhrase: map[string]types.MatchPhraseQuery{
			"ip": {Query: "192.168.120.0/24"},
		},
	}
	q4 := types.Query{
		MatchPhrase: map[string]types.MatchPhraseQuery{
			"org": {Query: "测试机构2"},
		},
	}
	query.Bool.Must = append(query.Bool.Must, q1)
	query.Bool.Must = append(query.Bool.Must, q4)

	assets := NewAssets("test1")
	res, err := assets.Search(query, 1, 20)
	if err != nil {
		t.Errorf("search error: %v", err)
		return
	}
	outputDoc(res, t)
}

func TestAssets_Aggregation2(t *testing.T) {
	workspaceResult := map[int]string{
		1: "b0c79065-7ff7-32ae-cc18-864ccd8f7717",
	}
	query := types.Query{
		MatchAll: types.NewMatchAllQuery(),
	}
	a := NewAssets(workspaceResult[1])
	result, err := a.Aggregation(query)
	if err != nil {
		t.Error(err)
		return
	}
	t.Log(utils.SortMapByValue(result["port"], true))
	t.Log(utils.SortMapByValue(result["icon_hash"], true))
	t.Log(utils.SortMapByValue(result["service"], true))
	t.Log(utils.SortMapByValue(result["server"], true))
	t.Log(utils.SortMapByValue(result["location"], true))
	t.Log(utils.SortMapByValue(result["title"], true))
}
