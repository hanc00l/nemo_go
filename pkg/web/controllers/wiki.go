package controllers

import (
	"fmt"
	"github.com/hanc00l/nemo_go/pkg/conf"
	"github.com/hanc00l/nemo_go/pkg/db"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"github.com/hanc00l/nemo_go/pkg/utils"
	"github.com/hanc00l/nemo_go/pkg/wiki"
	"os"
	"path"
	"regexp"
	"time"
)

type WikiController struct {
	BaseController
}

type DocumentInfo struct {
	Id             int    `json:"id"`
	Index          int    `json:"index"`
	Title          string `json:"title"`
	SpaceId        string `json:"space_id"`
	NodeToken      string `json:"node_token"`
	Comment        string `json:"comment"`
	PinIndex       int    `json:"pin_index"`
	ExportPathFile string `json:"export"`
	CreateTime     string `json:"create_time"`
	UpdateTime     string `json:"update_time"`
}

// FeishuAuthorizationCodeCallbackAction 飞书授权回调
func (c *WikiController) FeishuAuthorizationCodeCallbackAction() {
	defer c.ServeJSON()

	code := c.GetString("code")
	state := c.GetString("state")
	if !validAuthorizationCodeAndState(code, state) {
		c.FailedStatus("invalid code or state")
		return
	}
	feishu := wiki.NewFeishuWiki()
	if feishu.GetUserAccessTokenByCode(code) {
		c.SucceededStatus("获取用户AccessToken成功")
	} else {
		c.FailedStatus("获取用户AccessToken失败")
	}
}

// RefreshTokenAction 刷新用户访问Token
func (c *WikiController) RefreshTokenAction() {
	defer c.ServeJSON()
	if !c.CheckMultiAccessRequest([]RequestRole{SuperAdmin, Admin}, false) {
		c.FailedStatus("当前用户权限不允许！")
		return
	}

	feishu := wiki.NewFeishuWiki()
	c.MakeStatusResponse(feishu.RefreshUserAccessToken())
}

// IndexAction 列表页面
func (c *WikiController) IndexAction() {
	c.CheckMultiAccessRequest([]RequestRole{SuperAdmin, Admin}, true)

	c.Layout = "base.html"
	c.TplName = "wiki-list.html"
}

// ListAction 列表的数据
func (c *WikiController) ListAction() {
	defer c.ServeJSON()
	if !c.CheckMultiAccessRequest([]RequestRole{SuperAdmin, Admin}, false) {
		c.FailedStatus("当前用户权限不允许！")
		return
	}

	req := DatableRequestParam{}
	err := c.ParseForm(&req)
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
	}
	resp := c.getDocumentList(&req)
	c.Data["json"] = resp
}

// SyncFeishuSpacesAction 同步飞书知识库
func (c *WikiController) SyncFeishuSpacesAction() {
	defer c.ServeJSON()
	if !c.CheckMultiAccessRequest([]RequestRole{SuperAdmin, Admin}, false) {
		c.FailedStatus("当前用户权限不允许！")
		return
	}
	workspaceId := c.GetCurrentWorkspace()
	if workspaceId <= 0 {
		c.FailedStatus("当前工作空间不存在")
		return
	}
	spaceId := c.getCurrentSpaceId()
	if spaceId == "" {
		c.FailedStatus("当前工作空间没有设置知识库的spaceId")
		return
	}

	feishu := wiki.NewFeishuWiki()
	err, result := feishu.SyncWikiDocument(workspaceId, spaceId)
	if err != nil {
		c.FailedStatus(err.Error())
		return
	}
	c.SucceededStatus(result)
}

// AddIndexAction 新建一个文档页面
func (c *WikiController) AddIndexAction() {
	c.CheckMultiAccessRequest([]RequestRole{SuperAdmin, Admin}, true)
	c.Layout = "base.html"
	c.TplName = "wiki-add.html"
}

// AddSaveAction 新建一个文档
func (c *WikiController) AddSaveAction() {
	defer c.ServeJSON()
	if !c.CheckMultiAccessRequest([]RequestRole{SuperAdmin, Admin}, false) {
		c.FailedStatus("当前用户权限不允许！")
		return
	}

	spaceId := c.getCurrentSpaceId()
	if len(spaceId) == 0 {
		c.FailedStatus("当前工作空间没有设置知识库的spaceId")
		return
	}

	title := c.GetString("title", "")
	comment := c.GetString("comment", "")
	pinIndex, _ := c.GetInt("pin_index", 0)
	if title == "" {
		c.FailedStatus("title为空！")
		return
	}
	feishu := wiki.NewFeishuWiki()
	err, nodeToken := feishu.NewDocument(spaceId, title, comment, pinIndex)
	if err != nil {
		c.FailedStatus(err.Error())
		return
	}

	c.SucceededStatus(nodeToken)
}

// GetAction 根据token从数据库获取一个文档
func (c *WikiController) GetAction() {
	defer c.ServeJSON()
	if !c.CheckMultiAccessRequest([]RequestRole{SuperAdmin, Admin}, false) {
		c.FailedStatus("当前用户权限不允许！")
		return
	}

	id, err := c.GetInt("id", 0)
	if err != nil {
		c.FailedStatus(err.Error())
		return
	}
	docData := DocumentInfo{}
	docs := db.WikiDocs{Id: id}
	if docs.Get() {
		docData.Id = docs.Id
		docData.Title = docs.Title
		docData.Comment = docs.Comment
		docData.PinIndex = docs.PinIndex
	}
	c.Data["json"] = docData
}

// UpdateAction 更新一个数据库中的文档
func (c *WikiController) UpdateAction() {
	defer c.ServeJSON()
	if !c.CheckMultiAccessRequest([]RequestRole{SuperAdmin, Admin}, false) {
		c.FailedStatus("当前用户权限不允许！")
		return
	}
	id, err := c.GetInt("id", 0)
	if err != nil {
		c.FailedStatus(err.Error())
		return
	}
	if id <= 0 {
		c.FailedStatus("error id")
		return
	}
	docs := db.WikiDocs{Id: id}
	pinIndex, _ := c.GetInt("pin_index", 0)
	isRemoveExportedFile, _ := c.GetBool("remove_exported_file", false)
	// 只能更新数据库中的备注、置顶信息，其它文档信息由飞书同步
	if docs.Get() {
		searchMap := make(map[string]interface{})
		searchMap["comment"] = c.GetString("comment", "")
		searchMap["pin_index"] = pinIndex
		docs.Update(searchMap)
		// 清除导出的文档
		if isRemoveExportedFile {
			if workspaceGUID := c.GetCurrentWorkspaceGUID(); workspaceGUID != "" {
				// 删除导出的文档
				documentPathFile := path.Join(conf.GlobalServerConfig().Web.WebFiles, workspaceGUID, "wiki", fmt.Sprintf("%s.%s", docs.NodeToken, docs.ObjType))
				if utils.CheckFileExist(documentPathFile) {
					err = os.RemoveAll(documentPathFile)
					if err != nil {
						logging.RuntimeLog.Error(err)
						c.FailedStatus("删除导出的文档失败")
						return
					}
				}
			}
		}
		c.SucceededStatus("更新成功")
		return
	} else {
		c.FailedStatus("更新失败，文档不存在")
	}
}

// ExportDocument 导出一个文档
func (c *WikiController) ExportDocument() {
	defer c.ServeJSON()
	if !c.CheckMultiAccessRequest([]RequestRole{SuperAdmin, Admin}, false) {
		c.FailedStatus("当前用户权限不允许！")
		return
	}
	//
	id, err := c.GetInt("id", 0)
	if err != nil {
		c.FailedStatus(err.Error())
		return
	}
	if id <= 0 {
		c.FailedStatus("error id")
		return
	}
	//
	docs := db.WikiDocs{Id: id}
	if !docs.Get() {
		c.FailedStatus("文档不存在")
		return
	}
	feishu := wiki.NewFeishuWiki()
	// 创建导出任务
	err, ticket := feishu.CreateExportTask(docs.ObjType, docs.ObjToken)
	if err != nil {
		logging.RuntimeLog.Error(err)
		c.FailedStatus("创建导出任务失败")
		return
	}
	if ticket == "" {
		c.FailedStatus("创建导出任务失败")
		return
	}
	// 获取导出任务结果
	var fileToken string
	var d int
	for {
		if d++; d > 30 {
			c.FailedStatus("查询导出任务超时")
			return
		}
		time.Sleep(1 * time.Second)
		err, fileToken = feishu.QueryExportTask(docs.ObjToken, ticket)
		if err != nil {
			logging.RuntimeLog.Error(err)
			c.FailedStatus("查询导出任务失败")
			return
		}
		if fileToken != "" {
			break
		}
	}
	if fileToken == "" {
		c.FailedStatus("查询导出任务失败")
		return
	}
	//检查保存结果的路径
	workspaceGUID := c.GetCurrentWorkspaceGUID()
	if workspaceGUID == "" {
		c.FailedStatus("workspace error")
		return
	}
	wikiPath := path.Join(conf.GlobalServerConfig().Web.WebFiles, workspaceGUID, "wiki")
	if !utils.MakePath(wikiPath) {
		c.FailedStatus("创建保存wiki的目录失败！")
		return
	}
	// 下载导出的文档
	documentPathFile := path.Join(wikiPath, fmt.Sprintf("%s.%s", docs.NodeToken, docs.ObjType))
	err = feishu.DownloadExportTask(fileToken, documentPathFile)
	if err != nil {
		logging.RuntimeLog.Error(err)
		c.FailedStatus("下载文档失败")
		return
	}
	c.SucceededStatus("导出文档成功")
}

// getDocumentList 获取当前与工作空间相关联的知识库中的文档列表
func (c *WikiController) getDocumentList(req *DatableRequestParam) (resp DataTableResponseData) {
	spaceId := c.getCurrentSpaceId()
	if spaceId == "" {
		resp.Draw = req.Draw
		resp.RecordsTotal = 0
		resp.RecordsFiltered = 0
		resp.Data = make([]interface{}, 0)
		return
	}
	var wikiPath, wikiWebPath string
	workspaceGUID := c.GetCurrentWorkspaceGUID()
	if workspaceGUID != "" {
		wikiPath = path.Join(conf.GlobalServerConfig().Web.WebFiles, workspaceGUID, "wiki")
		wikiWebPath = path.Join("webfiles", workspaceGUID, "wiki")

	}
	docs := db.WikiDocs{}
	searchMap := make(map[string]interface{})
	searchMap["space_id"] = spaceId
	docsResult, total := docs.Gets(searchMap, req.Start/req.Length+1, req.Length)
	index := 1
	for _, item := range docsResult {
		d := DocumentInfo{}
		d.Id = item.Id
		d.Index = index
		d.Title = item.Title
		d.SpaceId = item.SpaceID
		d.NodeToken = item.NodeToken
		d.Comment = item.Comment
		d.PinIndex = item.PinIndex
		d.CreateTime = FormatDateTime(item.CreateDatetime)
		d.UpdateTime = FormatSubDateTime(item.UpdateDatetime)
		if wikiPath != "" {
			d.ExportPathFile = checkDocumentExportFile(wikiPath, wikiWebPath, item.NodeToken, item.ObjType)
		}
		resp.Data = append(resp.Data, d)
		index++
	}
	resp.Draw = req.Draw
	resp.RecordsTotal = total
	resp.RecordsFiltered = total
	if total <= 0 {
		resp.Data = make([]interface{}, 0)
	}

	return
}

// getCurrentSpaceId 根据当前的workspace获取spaceId
func (c *WikiController) getCurrentSpaceId() string {
	workspaceId := c.GetCurrentWorkspace()
	if workspaceId <= 0 {
		return ""
	}
	s := db.WikiSpace{WorkspaceId: workspaceId}
	if s.GetByWorkspaceId() {
		return s.WikiSpaceId
	}
	return ""
}

// FormatSubDateTime 统一格式化当前时间间隔
func FormatSubDateTime(dt time.Time) string {
	d := time.Now().Sub(dt).Truncate(time.Minute)
	if d.Hours() > 24*30 {
		return fmt.Sprintf("%d个月前", int(d.Hours()/24/30))
	} else if d.Hours() > 24 {
		return fmt.Sprintf("%d天前", int(d.Hours()/24))
	} else if d.Minutes() > 60 {
		return fmt.Sprintf("%d小时前", int(d.Hours()))
	} else if d.Minutes() > 1 {
		return fmt.Sprintf("%d分钟前", int(d.Minutes()))
	} else {
		return "刚刚"
	}
}

func validAuthorizationCodeAndState(code, state string) bool {
	if state != "RANDOMSTATE" {
		logging.CLILog.Errorf("invalid state:%s", state)
		return false
	}
	if len(code) != 32 {
		logging.CLILog.Errorf("invalid code:%s", code)
		return false
	}
	if !regexp.MustCompile(`^[a-zA-Z0-9]+$`).MatchString(code) {
		logging.CLILog.Errorf("invalid code:%s", code)
		return false
	}

	return true
}

func checkDocumentExportFile(wikiPath, wikiWebPath, nodeToken, objType string) (exportWebPathFile string) {
	documentPathFile := path.Join(wikiPath, fmt.Sprintf("%s.%s", nodeToken, objType))
	if utils.CheckFileExist(documentPathFile) {
		exportWebPathFile = path.Join(wikiWebPath, fmt.Sprintf("%s.%s", nodeToken, objType))
	}
	return
}
