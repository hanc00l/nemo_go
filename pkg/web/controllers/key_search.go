package controllers

import (
	"github.com/hanc00l/nemo_go/pkg/db"
	"github.com/hanc00l/nemo_go/pkg/logging"
)

type KeySearchController struct {
	BaseController
}
type keyWordInitRequestParam struct {
	AddOrgId        int    `form:"add_org_id"`
	AddKeyWord      string `form:"add_key_word"`
	AddSearchTime   string `form:"add_search_time"`
	AddExcludeWords string `form:"add_exclude_words"`
	AddCheckMod     string `form:"add_check_mod"`
	AddCount        int    `form:"add_count"`
}

type keySearchRequestParam struct {
	DatableRequestParam
	OrgId        int    `form:"org_id"`
	KeyWord      string `form:"key_word"`
	SearchTime   string `form:"search_time"`
	ExcludeWords string `form:"exclude_words"`
	CheckMod     string `form:"check_mod"`
}

type KeyWordList struct {
	Id             int    `json:"id"`
	Index          int    `json:"index"`
	OrgId          string `json:"org_id"`
	KeyWord        string `json:"key_word"`
	SearchTime     string `json:"search_time"`
	ExcludeWords   string `json:"exclude_words"`
	CheckMod       string `json:"check_mod"`
	IsDelete       bool   `json:"is_delete"`
	Count          int    `json:"count"`
	CreateDatetime string `json:"create_datetime"`
	UpdateDatetime string `json:"update_datetime"`
	WorkspaceId    int    `json:"workspace"`
}
type KeyWordInfo struct {
	Id             int    `json:"id"`
	OrgId          int    `json:"org_id"`
	KeyWord        string `json:"key_word"`
	SearchTime     string `json:"search_time"`
	ExcludeWords   string `json:"exclude_words"`
	CheckMod       string `json:"check_mod"`
	IsDelete       bool   `json:"is_delete"`
	Count          int    `json:"count"`
	CreateDatetime string `json:"create_datetime"`
	UpdateDatetime string `json:"update_datetime"`
}

func (c *KeySearchController) IndexAction() {
	c.UpdateOnlineUser()
	c.Layout = "base.html"
	c.TplName = "key-word-list.html"
}

func (c *KeySearchController) IndexBlackAction() {
	c.UpdateOnlineUser()
	c.Layout = "base.html"
	c.TplName = "task-cron-list.html"
}

// AddSaveAction 保存新增的记录
func (c *KeySearchController) AddSaveAction() {
	defer c.ServeJSON()
	if c.CheckMultiAccessRequest([]RequestRole{SuperAdmin, Admin}, false) == false {
		c.FailedStatus("当前用户权限不允许！")
		return
	}

	workspaceId := c.GetSession("Workspace").(int)
	if workspaceId <= 0 {
		c.FailedStatus("未选择当前的工作空间！")
		return
	}
	keyWordData := keyWordInitRequestParam{}
	err := c.ParseForm(&keyWordData)
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		c.FailedStatus(err.Error())
		return
	}
	kw := db.KeyWord{}
	kw.OrgId = keyWordData.AddOrgId
	kw.KeyWord = keyWordData.AddKeyWord
	kw.SearchTime = keyWordData.AddSearchTime
	kw.ExcludeWords = keyWordData.AddExcludeWords
	kw.CheckMod = keyWordData.AddCheckMod
	kw.Count = keyWordData.AddCount
	kw.WorkspaceId = workspaceId
	c.MakeStatusResponse(kw.Add())
}

// validateRequestParam 校验请求的参数
func (c *KeySearchController) validateRequestParam(req *keySearchRequestParam) {
	if req.Length <= 0 {
		req.Length = 50
	}
	if req.Start < 0 {
		req.Start = 0
	}
}

// ListAction IP列表
func (c *KeySearchController) ListAction() {
	c.UpdateOnlineUser()
	defer c.ServeJSON()

	req := keySearchRequestParam{}
	err := c.ParseForm(&req)
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
	}
	c.validateRequestParam(&req)

	resp := c.getKeyWordListData(req)
	c.Data["json"] = resp
}

// DeleteKeyWordAction 删除一个记录
func (c *KeySearchController) DeleteKeyWordAction() {
	defer c.ServeJSON()
	if c.CheckMultiAccessRequest([]RequestRole{SuperAdmin, Admin}, false) == false {
		c.FailedStatus("当前用户权限不允许！")
		return
	}

	id, err := c.GetInt("id")
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		c.FailedStatus(err.Error())
		return
	}
	ip := db.KeyWord{Id: id}
	updateMap := make(map[string]interface{})
	updateMap["is_delete"] = true
	ip.Update(updateMap)
	c.MakeStatusResponse(true)
}

// getSearchMap 根据查询参数生成查询条件
func (c *KeySearchController) getSearchMap(req keySearchRequestParam) (searchMap map[string]interface{}) {
	searchMap = make(map[string]interface{})

	workspaceId := c.GetSession("Workspace").(int)
	if workspaceId > 0 {
		searchMap["workspace_id"] = workspaceId
	}
	if req.KeyWord != "" {
		searchMap["key_word"] = req.KeyWord
	}
	if req.SearchTime != "" {
		searchMap["search_time"] = req.SearchTime
	}
	if req.ExcludeWords != "" {
		searchMap["exclude_words"] = req.ExcludeWords
	}
	if req.CheckMod != "" {
		searchMap["check_mod"] = req.CheckMod
	}
	if req.OrgId > 0 {
		searchMap["org_id"] = req.OrgId
	}
	return
}

// getKeyWordListData 获取列表数据
func (c *KeySearchController) getKeyWordListData(req keySearchRequestParam) (resp DataTableResponseData) {
	keyWords := db.KeyWord{}
	searchMap := c.getSearchMap(req)
	workspaceId := c.GetSession("Workspace").(int)
	if workspaceId > 0 {
		searchMap["workspace_id"] = workspaceId
	}
	startPage := req.Start/req.Length + 1
	results, total := keyWords.Gets(searchMap, startPage, req.Length)
	for i, keyWordRow := range results {
		r := KeyWordList{}
		r.Id = keyWordRow.Id
		r.KeyWord = keyWordRow.KeyWord
		r.Index = req.Start + i + 1
		r.CheckMod = keyWordRow.CheckMod
		r.ExcludeWords = keyWordRow.ExcludeWords
		r.SearchTime = keyWordRow.SearchTime
		r.WorkspaceId = keyWordRow.WorkspaceId
		orgDb := db.Organization{}
		orgDb.Id = keyWordRow.OrgId
		if orgDb.Get() {
			r.OrgId = orgDb.OrgName
		}
		r.Count = keyWordRow.Count
		r.UpdateDatetime = FormatDateTime(keyWordRow.UpdateDatetime)
		r.CreateDatetime = FormatDateTime(keyWordRow.CreateDatetime)

		resp.Data = append(resp.Data, r)
	}
	resp.Draw = req.Draw
	resp.RecordsTotal = total
	resp.RecordsFiltered = total
	if resp.Data == nil {
		resp.Data = make([]interface{}, 0)
	}
	return resp
}
