package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/elastic/go-elasticsearch/v8/typedapi/types"
	"github.com/hanc00l/nemo_go/v2/pkg/conf"
	"github.com/hanc00l/nemo_go/v2/pkg/es"
	"github.com/hanc00l/nemo_go/v2/pkg/logging"
	"github.com/hanc00l/nemo_go/v2/pkg/utils"
	"path/filepath"
	"strings"
)

type EsSearchController struct {
	BaseController
}

// AssetsListData datable显示的每一行数据
type AssetsListData struct {
	Id             string   `json:"id"`
	Index          int      `json:"index"`
	Host           string   `json:"host"`
	Domain         string   `json:"domain"`
	IP             []string `json:"ip"`
	Port           string   `json:"port"`
	Status         string   `json:"status"`
	Location       []string `json:"location"`
	Service        string   `json:"service"`
	IconHash       int64    `json:"icon_hash"`
	Title          string   `json:"title"`
	Header         string   `json:"header"`
	Cert           string   `json:"cert"`
	Banner         string   `json:"banner"`
	WorkspaceId    int      `json:"workspace"`
	WorkspaceGUID  string   `json:"workspace_guid"`
	ScreenshotFile []string `json:"screenshot"`
	IconImage      []string `json:"iconimage"`
	UpdateTime     string   `json:"update_time"`
}

// esRequestParam 请求参数
type esRequestParam struct {
	DatableRequestParam
	Query string `form:"query"`
}

func (c *EsSearchController) IndexAction() {
	c.Layout = "base.html"
	if es.CheckElasticConn() {
		c.TplName = "es-list.html"
	} else {
		c.TplName = "es-list-empty.html"
	}
}

// ListAction 列表的数据
func (c *EsSearchController) ListAction() {
	defer c.ServeJSON()

	req := esRequestParam{}
	err := c.ParseForm(&req)
	if err != nil {
		logging.RuntimeLog.Error(err)
	}
	c.validateRequestParam(&req)
	resp := c.getEsListData(req)
	if resp.Data == nil {
		resp.Data = make([]interface{}, 0)
	}
	c.Data["json"] = resp
}

// DeleteAction 删除一个记录
func (c *EsSearchController) DeleteAction() {
	defer c.ServeJSON()
	if c.CheckMultiAccessRequest([]RequestRole{SuperAdmin, Admin}, false) == false {
		c.FailedStatus("当前用户权限不允许！")
		return
	}
	id := c.GetString("id", "")
	if id == "" {
		logging.RuntimeLog.Error("获取id失败")
		c.FailedStatus("获取id失败")
		return
	}
	workspace := c.GetCurrentWorkspaceGUID()
	if workspace == "" {
		c.FailedStatus("获取workspace错误")
		return
	}
	assets := es.NewAssets(workspace)
	c.MakeStatusResponse(assets.DeleteDoc(id))
}

// GetAssetsBody 获取资产的body信息
func (c *EsSearchController) GetAssetsBody() {
	defer c.ServeJSON()

	workspace := c.GetString("workspace", "")
	id := c.GetString("id", "")
	if id == "" || workspace == "" {
		logging.RuntimeLog.Error("获取参数失败")
		c.FailedStatus("获取参数失败")
		return
	}
	assets := es.NewAssets(workspace)
	if doc, ok := assets.GetDoc(id); !ok {
		c.FailedStatus("获取文档失败")
	} else {
		c.SucceededStatus(doc.Body)
	}
}

// validateRequestParam 校验请求的参数
func (c *EsSearchController) validateRequestParam(req *esRequestParam) {
	if req.Length <= 0 {
		req.Length = 50
	}
	if req.Start < 0 {
		req.Start = 0
	}
}

// getEsListData 根据查询条件获取数据
func (c *EsSearchController) getEsListData(req esRequestParam) (resp DataTableResponseData) {
	resp.Draw = req.Draw
	// 获取workspace
	workspaceId := c.GetCurrentWorkspace()
	workspaceGUID := c.GetCurrentWorkspaceGUID()
	if workspaceId <= 0 || workspaceGUID == "" {
		return
	}
	var err error
	var query types.Query
	// 如果查询条件为空，则默认查询所有
	if strings.TrimSpace(req.Query) == "" {
		query = types.Query{
			MatchAll: &types.MatchAllQuery{},
		}
	} else {
		query, err = es.ParseQuery(req.Query)
		if err != nil {
			logging.RuntimeLog.Errorf("parse query error: %v", err)
			return
		}
	}
	a := es.NewAssets(workspaceGUID)
	res, err := a.Search(query, req.Start/req.Length+1, req.Length)
	if err != nil {
		logging.RuntimeLog.Errorf("search error: %v", err)
		return
	}
	for i, hit := range res.Hits.Hits {
		var doc es.Document
		err = json.Unmarshal(hit.Source_, &doc)
		if err != nil {
			logging.RuntimeLog.Errorf("unmarshal document error: %v", err)
			continue
		}
		assetsData := AssetsListData{
			Id:            doc.Id,
			Index:         req.Start + i + 1,
			Host:          doc.Host,
			Domain:        doc.Domain,
			IP:            doc.Ip,
			Location:      doc.Location,
			Service:       doc.Service,
			IconHash:      doc.IconHash,
			Title:         doc.Title,
			Header:        doc.Header,
			Cert:          doc.Cert,
			Banner:        doc.Banner,
			WorkspaceId:   workspaceId,
			WorkspaceGUID: workspaceGUID,
			UpdateTime:    FormatDateTime(doc.UpdateTime),
		}
		if doc.Port > 0 {
			assetsData.Port = fmt.Sprintf("%d", doc.Port)
		}
		if doc.Status > 0 {
			assetsData.Status = fmt.Sprintf("%d", doc.Status)
		}
		assetsData.IconImage = loadIconImageFile(workspaceGUID, doc.IconHash)
		assetsData.ScreenshotFile = loadScreenshotFile(workspaceGUID, doc.Host)

		resp.Data = append(resp.Data, assetsData)
	}
	resp.RecordsTotal = int(res.Hits.Total.Value)
	resp.RecordsFiltered = resp.RecordsTotal
	return
}

func loadIconImageFile(workspaceGUID string, iconHash int64) (iconFiles []string) {
	imageFileBase := fmt.Sprintf("%s", utils.MD5(fmt.Sprintf("%d", iconHash)))
	files, _ := filepath.Glob(filepath.Join(conf.GlobalServerConfig().Web.WebFiles, workspaceGUID, "iconimage", fmt.Sprintf("%s.*", imageFileBase)))
	for _, file := range files {
		_, f := filepath.Split(file)
		iconFiles = append(iconFiles, fmt.Sprintf("%d|%s", iconHash, f))
	}

	return
}

// LoadScreenshotFile 获取screenshot文件
func loadScreenshotFile(workspaceGUID, host string) (r []string) {
	hostArray := strings.Split(host, ":")

	if len(hostArray) == 1 {
		// domain
		files, _ := filepath.Glob(filepath.Join(conf.GlobalServerConfig().Web.WebFiles, workspaceGUID, "screenshot", hostArray[0], "*.png"))
		for _, file := range files {
			_, f := filepath.Split(file)
			if !strings.HasSuffix(f, "_thumbnail.png") {
				r = append(r, f)
			}
		}
	} else if len(hostArray) >= 2 {
		// ip:port
		hostIpv46 := strings.Join(hostArray[:len(hostArray)-1], ":")
		files, _ := filepath.Glob(filepath.Join(conf.GlobalServerConfig().Web.WebFiles, workspaceGUID, "screenshot", hostArray[0], fmt.Sprintf("%s_*.png", hostIpv46)))
		for _, file := range files {
			_, f := filepath.Split(file)
			if !strings.HasSuffix(f, "_thumbnail.png") {
				r = append(r, f)
			}
		}
	}
	return
}
