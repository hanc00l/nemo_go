package wiki

import (
	"context"
	"fmt"
	"github.com/hanc00l/nemo_go/v2/pkg/conf"
	"github.com/hanc00l/nemo_go/v2/pkg/db"
	"github.com/hanc00l/nemo_go/v2/pkg/logging"
	"github.com/hanc00l/nemo_go/v2/pkg/utils"
	lark "github.com/larksuite/oapi-sdk-go/v3"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkauthen "github.com/larksuite/oapi-sdk-go/v3/service/authen/v1"
	larkdocx "github.com/larksuite/oapi-sdk-go/v3/service/docx/v1"
	larkdrive "github.com/larksuite/oapi-sdk-go/v3/service/drive/v1"
	larkwiki "github.com/larksuite/oapi-sdk-go/v3/service/wiki/v2"
	"gopkg.in/errgo.v2/fmt/errors"
	"strconv"
	"sync"
	"time"
)

type FeishuWiki struct {
	appId                  string
	appSecret              string
	userAccessToken        string
	userAccessRefreshToken string
	tokenRefreshTime       *time.Time
	tokenMutex             sync.Mutex
}

var feishuWiki *FeishuWiki

// NewFeishuWiki 创建全局的feishu对象
func NewFeishuWiki() *FeishuWiki {
	if feishuWiki != nil {
		if feishuWiki.appId != "" && feishuWiki.appSecret != "" && feishuWiki.userAccessRefreshToken != "" {
			if feishuWiki.tokenRefreshTime == nil || time.Now().Sub(*feishuWiki.tokenRefreshTime) > 1*time.Hour {
				logging.RuntimeLog.Info("refresh user access token as maybe token timeout")
				if feishuWiki.RefreshUserAccessToken() {
					logging.RuntimeLog.Info("refresh user access token success")
					logging.CLILog.Info("refresh user access token success")
				} else {
					logging.RuntimeLog.Warning("refresh user access token fail")
					logging.CLILog.Warning("refresh user access token fail")
				}
			}
		}
		return feishuWiki
	}

	feishuWiki = &FeishuWiki{
		appId:                  conf.GlobalServerConfig().Wiki.Feishu.AppId,
		appSecret:              conf.GlobalServerConfig().Wiki.Feishu.AppSecret,
		userAccessRefreshToken: conf.GlobalServerConfig().Wiki.Feishu.UserAccessRefreshToken,
	}
	// 每次启动初始化时，刷新一次用户AccessToken
	if feishuWiki.appId != "" && feishuWiki.appSecret != "" && feishuWiki.userAccessRefreshToken != "" {
		if feishuWiki.RefreshUserAccessToken() {
			logging.RuntimeLog.Info("refresh user access token success")
			logging.CLILog.Info("refresh user access token success")
		} else {
			logging.RuntimeLog.Warning("refresh user access token fail")
			logging.CLILog.Warning("refresh user access token fail")
		}
	}
	return feishuWiki
}

// GetUserAccessTokenByCode 根据登录预授权码code获取用户AccessToken
// 登录预授权码code由Server的WikiController的回调函数获取，coe只能使用一次，获取用户AccessToken后，需要保存refreshToken
func (w *FeishuWiki) GetUserAccessTokenByCode(code string) (status bool) {
	// 创建 Client
	var client = lark.NewClient(w.appId, w.appSecret) //, lark.WithLogLevel(larkcore.LogLevelDebug))
	// 创建请求对象
	req := larkauthen.NewCreateOidcAccessTokenReqBuilder().
		Body(larkauthen.NewCreateOidcAccessTokenReqBodyBuilder().
			GrantType(`authorization_code`).
			Code(code).
			Build()).
		Build()
	// 发起请求
	resp, err := client.Authen.OidcAccessToken.Create(context.Background(), req)
	// 处理错误
	if err != nil {
		logging.RuntimeLog.Error(err)
		logging.CLILog.Error(err)
		return
	}
	// 服务端错误处理
	if !resp.Success() {
		logging.RuntimeLog.Error(resp.Code, resp.Msg, resp.RequestId())
		logging.CLILog.Error(resp.Code, resp.Msg, resp.RequestId())
		return
	}
	// 业务处理
	if resp.Data.AccessToken != nil && resp.Data.RefreshToken != nil {
		w.tokenMutex.Lock()
		defer w.tokenMutex.Unlock()

		status = true
		w.userAccessToken = *resp.Data.AccessToken
		w.userAccessRefreshToken = *resp.Data.RefreshToken
		dtNow := time.Now()
		w.tokenRefreshTime = &dtNow
		w.saveRefreshToken()
	}
	return
}

// RefreshUserAccessToken 根据refreshToken刷新用户AccessToken
// 用户AccessToken一般只有2小时有效期，而RefreshToken有30天有效期
func (w *FeishuWiki) RefreshUserAccessToken() (status bool) {
	// 创建 Client
	var client = lark.NewClient(w.appId, w.appSecret)
	// 创建请求对象
	req := larkauthen.NewCreateOidcRefreshAccessTokenReqBuilder().
		Body(larkauthen.NewCreateOidcRefreshAccessTokenReqBodyBuilder().
			GrantType(`refresh_token`).
			RefreshToken(w.userAccessRefreshToken).
			Build()).
		Build()
	// 发起请求
	resp, err := client.Authen.OidcRefreshAccessToken.Create(context.Background(), req)
	// 处理错误
	if err != nil {
		logging.RuntimeLog.Error(err)
		logging.CLILog.Error(err)
		return
	}
	// 服务端错误处理
	if !resp.Success() {
		logging.RuntimeLog.Error(resp.Code, resp.Msg, resp.RequestId())
		logging.CLILog.Error(resp.Code, resp.Msg, resp.RequestId())
		return
	}
	// 业务处理
	if resp.Data.AccessToken != nil && resp.Data.RefreshToken != nil {
		w.tokenMutex.Lock()
		defer w.tokenMutex.Unlock()

		status = true
		w.userAccessToken = *resp.Data.AccessToken
		w.userAccessRefreshToken = *resp.Data.RefreshToken
		dtNow := time.Now()
		w.tokenRefreshTime = &dtNow
		w.saveRefreshToken()
	}
	return
}

// GetDocuments 从飞书获取指定空间的文档列表
func (w *FeishuWiki) GetDocuments(spaceId string) (errResult error, results []db.WikiDocs) {
	// 创建 Client
	var client = lark.NewClient(w.appId, w.appSecret) //, lark.WithLogLevel(larkcore.LogLevelDebug))
	pageToken := ""
	for {
		// 创建请求对象
		req := larkwiki.NewListSpaceNodeReqBuilder().
			SpaceId(spaceId).
			PageSize(20).PageToken(pageToken).
			Build()
		// 发起请求
		resp, err := client.Wiki.V2.SpaceNode.List(context.Background(), req, larkcore.WithUserAccessToken(w.userAccessToken))
		// 处理错误
		if err != nil {
			logging.RuntimeLog.Error(err)
			logging.CLILog.Error(err)
			errResult = err
			return
		}
		// 服务端错误处理
		if !resp.Success() {
			errResult = errors.Newf("resp.Code:%d,resp.Msg:%s,resp.RequestId():%s", resp.Code, resp.Msg, resp.RequestId())
			logging.RuntimeLog.Error(errResult)
			logging.CLILog.Error(errResult)
			return
		}
		// 业务处理
		for _, item := range resp.Data.Items {
			doc := db.WikiDocs{
				SpaceID:   *item.SpaceId,
				NodeToken: *item.NodeToken,
				Title:     *item.Title,
				ObjType:   *item.ObjType,
				ObjToken:  *item.ObjToken,
			}
			nodeCreateTime, _ := strconv.Atoi(*item.NodeCreateTime)
			createTime := time.Unix(int64(nodeCreateTime), 0)
			doc.CreateDatetime = createTime
			nodeEditTime, _ := strconv.Atoi(*item.ObjEditTime)
			editTime := time.Unix(int64(nodeEditTime), 0)
			doc.UpdateDatetime = editTime
			results = append(results, doc)
		}
		if *resp.Data.HasMore {
			pageToken = *resp.Data.PageToken
		} else {
			break
		}
	}
	return
}

// GetDocument 根据文档Token获取文档内容
func (w *FeishuWiki) GetDocument(documentToken string) (errResult error, doc db.WikiDocs) {
	// 创建 Client
	var client = lark.NewClient(w.appId, w.appSecret)
	// 创建请求对象
	req := larkwiki.NewGetNodeSpaceReqBuilder().
		Token(documentToken).
		Build()
	// 发起请求
	resp, err := client.Wiki.Space.GetNode(context.Background(), req, larkcore.WithUserAccessToken(w.userAccessToken))
	// 处理错误
	if err != nil {
		logging.RuntimeLog.Error(err)
		logging.CLILog.Error(err)
		errResult = err
		return
	}
	// 服务端错误处理
	if !resp.Success() {
		errResult = errors.Newf("resp.Code:%d,resp.Msg:%s,resp.RequestId():%s", resp.Code, resp.Msg, resp.RequestId())
		logging.RuntimeLog.Error(errResult)
		logging.CLILog.Error(errResult)
		return
	}
	// 业务处理
	doc = db.WikiDocs{
		SpaceID:   *resp.Data.Node.SpaceId,
		NodeToken: *resp.Data.Node.NodeToken,
		Title:     *resp.Data.Node.Title,
		ObjType:   *resp.Data.Node.ObjType,
		ObjToken:  *resp.Data.Node.ObjToken,
	}
	nodeCreateTime, _ := strconv.Atoi(*resp.Data.Node.NodeCreateTime)
	createTime := time.Unix(int64(nodeCreateTime), 0)
	doc.CreateDatetime = createTime
	nodeEditTime, _ := strconv.Atoi(*resp.Data.Node.ObjEditTime)
	editTime := time.Unix(int64(nodeEditTime), 0)
	doc.UpdateDatetime = editTime

	return
}

// GetDocumentContent 获取文档的纯内容，用于分析文档内容
// 应用频率限制：单个应用调用频率上限为每秒 5 次，超过该频率限制，接口将返回 HTTP 状态码 400 及错误码 99991400。
// 当请求被限频，应用需要处理限频状态码，并使用指数退避算法或其它一些频控策略降低对 API 的调用速率。
func (w *FeishuWiki) GetDocumentContent(documentToken string) (errResult error, content string) {
	// 创建 Client
	var client = lark.NewClient(w.appId, w.appSecret)
	// 创建请求对象
	req := larkdocx.NewRawContentDocumentReqBuilder().
		DocumentId(documentToken).
		Build()
	// 发起请求
	resp, err := client.Docx.Document.RawContent(context.Background(), req, larkcore.WithUserAccessToken(w.userAccessToken))
	// 处理错误
	if err != nil {
		logging.RuntimeLog.Error(err)
		logging.CLILog.Error(err)
		errResult = err
		return
	}
	// 服务端错误处理
	if !resp.Success() {
		errResult = errors.Newf("resp.Code:%d,resp.Msg:%s,resp.RequestId():%s", resp.Code, resp.Msg, resp.RequestId())
		logging.RuntimeLog.Error(errResult)
		logging.CLILog.Error(errResult)
		return
	}
	// 业务处理
	//fmt.Println(larkcore.Prettify(resp))
	content = *resp.Data.Content

	return
}

// NewDocument 创建新的文档
func (w *FeishuWiki) NewDocument(spaceId, title, comment string, pinIndex int) (errResult error, nodeToken string) {
	// 创建 Client
	var client = lark.NewClient(w.appId, w.appSecret)
	// 创建请求对象
	req := larkwiki.NewCreateSpaceNodeReqBuilder().
		SpaceId(spaceId).
		Node(larkwiki.NewNodeBuilder().
			ObjType(`docx`).
			NodeType(`origin`).
			Title(title).
			Build()).
		Build()
	// 发起请求
	resp, err := client.Wiki.SpaceNode.Create(context.Background(), req, larkcore.WithUserAccessToken(w.userAccessToken))
	// 处理错误
	if err != nil {
		logging.RuntimeLog.Error(err)
		logging.CLILog.Error(err)
		errResult = err
		return
	}
	// 服务端错误处理
	if !resp.Success() {
		errResult = errors.Newf("resp.Code:%d,resp.Msg:%s,resp.RequestId():%s", resp.Code, resp.Msg, resp.RequestId())
		logging.RuntimeLog.Error(errResult)
		logging.CLILog.Error(errResult)
		return
	}
	// 业务处理
	nodeToken = *resp.Data.Node.NodeToken
	//写入到数据库中：
	docs := db.WikiDocs{
		SpaceID:        spaceId,
		NodeToken:      nodeToken,
		Title:          title,
		Comment:        comment,
		PinIndex:       pinIndex,
		ObjType:        *resp.Data.Node.ObjType,
		ObjToken:       *resp.Data.Node.ObjToken,
		CreateDatetime: time.Now(),
		UpdateDatetime: time.Now(),
	}
	if docs.Add() {
		return nil, nodeToken
	} else {
		return errors.New("save to nemo db fail"), nodeToken
	}
}

// SyncWikiDocument 从飞书平台同步文档到数据库中
func (w *FeishuWiki) SyncWikiDocument(workspaceId int, spaceId string) (err error, result string) {
	// 从飞书获取文档列表
	var docsFeishu []db.WikiDocs
	err, docsFeishu = w.GetDocuments(spaceId)
	if err != nil {
		return
	}
	docsExisted := db.WikiDocs{}
	docsResultExisted, _ := docsExisted.Gets(make(map[string]interface{}), -1, 100000)
	var docsUpdated []db.WikiDocs
	var docsNewNum, docsUpdatedNum, docsRemovedNum int
	var ipRelatedNum, domainRelatedNum int
	//更新或新增文档到数据库中
	for _, item := range docsFeishu {
		if exist, d := w.isDocumentExisted(docsResultExisted, item.NodeToken); exist {
			// 同步更新title和最后编辑时间
			updateMap := map[string]interface{}{}
			updateMap["title"] = item.Title
			updateMap["update_datetime"] = item.UpdateDatetime
			// 更新
			d.Update(updateMap)
			item.Id = d.Id
			docsUpdated = append(docsUpdated, d)
			docsUpdatedNum++
		} else {
			// 新增
			item.Add()
			fmt.Print("---", item.Id)
			docsNewNum++
		}
		// 提取关联信息：IP、域名
		ipNum, domainNum := w.findIPAndDomainInDocument(item.ObjToken, item.Id, workspaceId)
		ipRelatedNum += ipNum
		domainRelatedNum += domainNum
		// 由于飞书的API限制，每秒只能调用5次，所以每次调用后，休眠200毫秒
		time.Sleep(200 * time.Millisecond)
	}
	//从数据库移除已经不存在的文档
	for _, item := range docsResultExisted {
		if exist, _ := w.isDocumentExisted(docsUpdated, item.NodeToken); !exist {
			item.Delete()
			docsRemovedNum++
		}
	}
	result = fmt.Sprintf("同步完成：新增%d个文档，更新%d个文档，移除%d个文档；关联%d个IP，%d个域名", docsNewNum, docsUpdatedNum, docsRemovedNum, ipRelatedNum, domainRelatedNum)
	return
}

// findIPAndDomainInDocument 从文档内容中提取IP和域名，并保存到数据库中
func (w *FeishuWiki) findIPAndDomainInDocument(objToken string, docId int, workspaceId int) (ipRelated, domainRelate int) {
	// 清除原来的记录
	docIpOld := db.WikiDocsIP{}
	docIpOld.RemoveByDocId(docId)
	docDomainOld := db.WikiDocsDomain{}
	docDomainOld.RemoveByDocument(docId)
	// 获取文档内容
	err, content := w.GetDocumentContent(objToken)
	if err != nil || content == "" {
		return
	}
	// 解析文档内容，提取IP和域名
	for _, ipName := range utils.FindIPV4(content) {
		ip := db.Ip{WorkspaceId: workspaceId, IpName: ipName}
		if ip.GetByIp() {
			docIp := db.WikiDocsIP{IpId: ip.Id, DocumentId: docId}
			docIp.Add()
			ipRelated++
		}
	}
	for _, domainName := range utils.FindDomain(content) {
		domain := db.Domain{WorkspaceId: workspaceId, DomainName: domainName}
		if domain.GetByDomain() {
			docDomain := db.WikiDocsDomain{DomainId: domain.Id, DocumentId: docId}
			docDomain.Add()
			domainRelate++
		}
	}

	return
}

// CreateExportTask 创建指定文档的导出任务
func (w *FeishuWiki) CreateExportTask(objType, objToken string) (errResult error, ticket string) {
	// 创建 Client
	var client = lark.NewClient(w.appId, w.appSecret)
	// 创建请求对象
	req := larkdrive.NewCreateExportTaskReqBuilder().
		ExportTask(larkdrive.NewExportTaskBuilder().
			FileExtension(objType).
			Token(objToken).
			Type(objType).
			Build()).
		Build()
	// 发起请求
	resp, err := client.Drive.ExportTask.Create(context.Background(), req, larkcore.WithUserAccessToken(w.userAccessToken))
	// 处理错误
	if err != nil {
		logging.RuntimeLog.Error(err)
		logging.CLILog.Error(err)
		errResult = err
		return
	}
	// 服务端错误处理
	if !resp.Success() {
		errResult = errors.Newf("resp.Code:%d,resp.Msg:%s,resp.RequestId():%s", resp.Code, resp.Msg, resp.RequestId())
		logging.RuntimeLog.Error(errResult)
		logging.CLILog.Error(errResult)
		return
	}
	// 业务处理
	//fmt.Println(larkcore.Prettify(resp))
	ticket = *resp.Data.Ticket

	return
}

// QueryExportTask 查询导出任务的状态
func (w *FeishuWiki) QueryExportTask(objToken, ticket string) (errResult error, fileToken string) {
	// 创建 Client
	var client = lark.NewClient(w.appId, w.appSecret)
	// 创建请求对象
	req := larkdrive.NewGetExportTaskReqBuilder().
		Ticket(ticket).
		Token(objToken).
		Build()
	// 发起请求
	resp, err := client.Drive.ExportTask.Get(context.Background(), req, larkcore.WithUserAccessToken(w.userAccessToken))
	// 处理错误
	if err != nil {
		logging.RuntimeLog.Error(err)
		logging.CLILog.Error(err)
		errResult = err
		return
	}
	// 服务端错误处理
	if !resp.Success() {
		errResult = errors.Newf("resp.Code:%d,resp.Msg:%s,resp.RequestId():%s", resp.Code, resp.Msg, resp.RequestId())
		logging.RuntimeLog.Error(errResult)
		logging.CLILog.Error(errResult)
		return
	}
	// 业务处理
	//fmt.Println(larkcore.Prettify(resp))
	fileToken = *resp.Data.Result.FileToken

	return
}

// DownloadExportTask 下载导出任务的结果
func (w *FeishuWiki) DownloadExportTask(fileToken string, outputPathFile string) (errResult error) {
	// 创建 Client
	var client = lark.NewClient(w.appId, w.appSecret)
	// 创建请求对象
	req := larkdrive.NewDownloadExportTaskReqBuilder().
		FileToken(fileToken).
		Build()
	// 发起请求
	resp, err := client.Drive.ExportTask.Download(context.Background(), req, larkcore.WithUserAccessToken(w.userAccessToken))
	// 处理错误
	if err != nil {
		logging.RuntimeLog.Error(err)
		logging.CLILog.Error(err)
		errResult = err
		return
	}
	// 服务端错误处理
	if !resp.Success() {
		errResult = errors.Newf("resp.Code:%d,resp.Msg:%s,resp.RequestId():%s", resp.Code, resp.Msg, resp.RequestId())
		logging.RuntimeLog.Error(errResult)
		logging.CLILog.Error(errResult)
		return
	}
	// 业务处理
	errResult = resp.WriteFile(outputPathFile)

	return
}

// saveRefreshToken 保存refreshToken到文件中
func (w *FeishuWiki) saveRefreshToken() {
	// 保存refreshToken到配置文件
	err := conf.GlobalServerConfig().ReloadConfig()
	if err != nil {
		logging.RuntimeLog.Error("read config file error:", err)
		return
	}
	conf.GlobalServerConfig().Wiki.Feishu.UserAccessRefreshToken = w.userAccessRefreshToken
	err = conf.GlobalServerConfig().WriteConfig()
	if err != nil {
		logging.RuntimeLog.Error("save config file error:", err)
		return
	}
}

func (w *FeishuWiki) isDocumentExisted(docs []db.WikiDocs, nodeToken string) (bool, db.WikiDocs) {
	for _, item := range docs {
		if item.NodeToken == nodeToken {
			return true, item
		}
	}
	return false, db.WikiDocs{}
}
