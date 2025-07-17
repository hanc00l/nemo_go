package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/hanc00l/nemo_go/v3/pkg/conf"
	"github.com/hanc00l/nemo_go/v3/pkg/db"
	"github.com/hanc00l/nemo_go/v3/pkg/logging"
	"github.com/hanc00l/nemo_go/v3/pkg/task/execute"
	"github.com/hanc00l/nemo_go/v3/pkg/task/icp"
	"github.com/hanc00l/nemo_go/v3/pkg/task/llmapi"
	"github.com/hanc00l/nemo_go/v3/pkg/task/onlineapi"
	"github.com/remeh/sizedwaitgroup"
	"os"
	"path/filepath"
	"strings"
)

type ConfigController struct {
	BaseController
}

type WorkerConfigData struct {
	APIConfig    *conf.API    `json:"api,omitempty"`
	LLMAPIConfig *conf.LLMAPI `json:"llmapi,omitempty"`
	ProxyConfig  *conf.Proxy  `json:"proxy,omitempty"`
}

func (c *ConfigController) WorkerIndexAction() {
	c.CheckOneAccessRequest(SuperAdmin, true)
	c.Layout = "base.html"
	c.TplName = "config-worker.html"
}

func (c *ConfigController) ServerIndexAction() {
	c.Layout = "base.html"
	c.TplName = "config-server.html"
}

func (c *ConfigController) LoadWorkerConfigAction() {
	defer func(c *ConfigController, encoding ...bool) {
		_ = c.ServeJSON(encoding...)
	}(c)
	if !c.CheckOneAccessRequest(SuperAdmin, false) {
		c.FailedStatus("没有权限")
		return
	}
	globalConfig := conf.GlobalWorkerConfig()
	if globalConfig == nil {
		c.FailedStatus("配置不存在")
		return
	}
	err := globalConfig.ReloadConfig()
	if err != nil {
		c.FailedStatus(err.Error())
		return
	}
	config := WorkerConfigData{
		APIConfig:    &globalConfig.API,
		LLMAPIConfig: &globalConfig.LLMAPI,
		ProxyConfig:  &globalConfig.Proxy,
	}
	c.Data["json"] = &config
}

func (c *ConfigController) SaveWorkerConfigAction() {
	defer func(c *ConfigController, encoding ...bool) {
		_ = c.ServeJSON(encoding...)
	}(c)
	if !c.CheckOneAccessRequest(SuperAdmin, false) {
		c.FailedStatus("没有权限")
		return
	}
	var config WorkerConfigData
	err := json.Unmarshal(c.Ctx.Input.RequestBody, &config)
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		c.FailedStatus("参数错误")
		return
	}
	c.saveWorkerConfig(config)
}

func (c *ConfigController) saveWorkerConfig(config WorkerConfigData) {
	globalConfig := conf.GlobalWorkerConfig()
	if globalConfig == nil {
		c.FailedStatus("配置不存在")
		return
	}
	err := globalConfig.ReloadConfig()
	if err != nil {
		c.FailedStatus(err.Error())
		return
	}
	if config.APIConfig != nil {
		globalConfig.API = *config.APIConfig
	}
	if config.LLMAPIConfig != nil {
		globalConfig.LLMAPI = *config.LLMAPIConfig
	}
	if config.ProxyConfig != nil {
		globalConfig.Proxy = *config.ProxyConfig
	}
	err = globalConfig.WriteConfig()
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		c.FailedStatus("保存失败")
		return
	}
	c.SucceededStatus("保存成功")
}

func (c *ConfigController) ChangePasswordAction() {
	defer func(c *ConfigController, encoding ...bool) {
		_ = c.ServeJSON(encoding...)
	}(c)
	op := c.GetString("oldpass", "")
	np := c.GetString("newpass", "")
	if op == "" || np == "" {
		c.FailedStatus("参数为空！")
		return
	}
	userName := c.GetCurrentUser()
	if userName == "" {
		c.FailedStatus("未登录！")
		return
	}
	mongoClient, err := db.GetClient()
	if err != nil {
		logging.RuntimeLog.Error(err)
		c.FailedStatus("数据库连接失败！")
		return
	}
	defer db.CloseClient(mongoClient)

	user := db.NewUser(mongoClient)
	doc, err := user.GetByName(userName)
	if err != nil || doc.Username != userName {
		logging.RuntimeLog.Error(err)
		c.FailedStatus("用户不存在！")
		return
	}
	if doc.Password != PasswordEncrypt(op) {
		c.FailedStatus("旧密码错误！")
		return
	}
	doc.Password = PasswordEncrypt(np)
	c.CheckErrorAndStatus(user.Update(doc.Id.Hex(), doc))
}

func (c *ConfigController) LoadServerVersionAction() {
	defer func(c *ConfigController, encoding ...bool) {
		_ = c.ServeJSON(encoding...)
	}(c)

	if fileContent, err1 := os.ReadFile(filepath.Join(conf.GetRootPath(), "version.txt")); err1 == nil {
		c.SucceededStatus(strings.TrimSpace(string(fileContent)))
	} else {
		c.FailedStatus("获取版本号失败！")
	}
}

// TestOnlineAPIKeyAction 在线测试API的key是否可用
func (c *ConfigController) TestOnlineAPIKeyAction() {
	defer func(c *ConfigController, encoding ...bool) {
		err := c.ServeJSON(encoding...)
		if err != nil {

		}
	}(c)

	if c.CheckMultiAccessRequest([]RequestRole{SuperAdmin, Admin}, false) == false {
		c.FailedStatus("当前用户权限不允许！")
		return
	}
	sb := strings.Builder{}
	swg := sizedwaitgroup.New(4)
	msgChan := make(chan string)
	done := make(chan struct{})
	go func() {
		for {
			select {
			case msg := <-msgChan:
				sb.WriteString(msg)
			case <-done:
				return
			}
		}
	}()

	apiKeys := conf.GlobalWorkerConfig().API
	if len(apiKeys.Fofa.Key) > 0 {
		swg.Add()
		go testOnlineAPI("fofa", &swg, msgChan)
	}
	if len(apiKeys.Hunter.Key) > 0 {
		swg.Add()
		go testOnlineAPI("hunter", &swg, msgChan)
	}
	if len(apiKeys.Quake.Key) > 0 {
		swg.Add()
		go testOnlineAPI("quake", &swg, msgChan)
	}
	if len(apiKeys.ICPChinaz.Key) > 0 {
		swg.Add()
		go testICPAPI("chinaz", &swg, msgChan)
	}
	if len(apiKeys.ICPPlusChinaz.Key) > 0 {
		swg.Add()
		go testICPPlusAPI("chinaz", &swg, msgChan)
	}
	if len(apiKeys.ICPBeianx.Key) > 0 {
		swg.Add()
		go testICPAPI("beianx", &swg, msgChan)
		swg.Add()
		go testICPPlusAPI("beianx", &swg, msgChan)
	}
	swg.Wait()
	done <- struct{}{}
	if sb.Len() > 0 {
		c.SucceededStatus(sb.String())
	} else {
		c.FailedStatus("api接口没有可用的key！")
	}
}

func (c *ConfigController) TestLLMAPIKeyAction() {
	defer func(c *ConfigController, encoding ...bool) {
		err := c.ServeJSON(encoding...)
		if err != nil {

		}
	}(c)

	if c.CheckMultiAccessRequest([]RequestRole{SuperAdmin, Admin}, false) == false {
		c.FailedStatus("当前用户权限不允许！")
		return
	}
	sb := strings.Builder{}
	swg := sizedwaitgroup.New(4)
	msgChan := make(chan string)
	done := make(chan struct{})
	go func() {
		for {
			select {
			case msg := <-msgChan:
				sb.WriteString(msg)
			case <-done:
				return
			}
		}
	}()

	apiKeys := conf.GlobalWorkerConfig().LLMAPI
	if len(apiKeys.Kimi.Token) > 0 {
		swg.Add()
		go testLLMApi("kimi", &swg, msgChan)
	}
	if len(apiKeys.DeepSeek.Token) > 0 {
		swg.Add()
		go testLLMApi("deepseek", &swg, msgChan)
	}
	if len(apiKeys.Qwen.Token) > 0 {
		swg.Add()
		go testLLMApi("qwen", &swg, msgChan)
	}
	swg.Wait()
	done <- struct{}{}
	if sb.Len() > 0 {
		c.SucceededStatus(sb.String())
	} else {
		c.FailedStatus("api接口没有可用的Token！")
	}
}

func testOnlineAPI(apiName string, swg *sizedwaitgroup.SizedWaitGroup, testMsgChan chan string) {
	defer swg.Done()

	executorConfig := execute.ExecutorConfig{
		OnlineAPI: map[string]execute.OnlineAPIConfig{
			"fofa":   {SearchLimitCount: 100, SearchPageSize: 100},
			"hunter": {SearchLimitCount: 100, SearchPageSize: 100},
			"quake":  {SearchLimitCount: 100, SearchPageSize: 100},
		},
	}
	taskInfo := execute.ExecutorTaskInfo{
		MainTaskInfo: execute.MainTaskInfo{
			WorkspaceId:    "test",
			Target:         "fofa.info",
			OrgId:          "test",
			MainTaskId:     "onlineapi_test",
			ExecutorConfig: executorConfig,
		},
	}
	taskInfo.MainTaskInfo.ExecutorConfig = executorConfig

	taskInfo.Executor = apiName
	result := onlineapi.Do(taskInfo)

	if len(result) > 0 {
		testMsgChan <- fmt.Sprintf("%s: OK!\n", apiName)
	} else {
		testMsgChan <- fmt.Sprintf("%s: fail\n", apiName)
	}
	return
}

// testOnlineAPI 多线程方式测试在线接口的可用性
func testICPAPI(apiName string, swg *sizedwaitgroup.SizedWaitGroup, testMsgChan chan string) {
	defer swg.Done()

	executorConfig := execute.ExecutorConfig{
		ICP: map[string]execute.ICPConfig{
			"icp": {APIName: []string{apiName}},
		},
	}
	taskInfo := execute.ExecutorTaskInfo{
		MainTaskInfo: execute.MainTaskInfo{
			WorkspaceId:    "test",
			Target:         "fofa.info",
			OrgId:          "test",
			MainTaskId:     "icp_test",
			ExecutorConfig: executorConfig,
		},
		Executor: "icp",
	}
	taskInfo.MainTaskInfo.ExecutorConfig = executorConfig
	result := icp.Do(taskInfo, true)
	if len(result) > 0 {
		testMsgChan <- fmt.Sprintf("%s-%s: OK!\n", "ICP备案查询", apiName)
	} else {
		testMsgChan <- fmt.Sprintf("%s-%s: fail\n", "ICP备案查询", apiName)
	}
	return
}

func testICPPlusAPI(apiName string, swg *sizedwaitgroup.SizedWaitGroup, testMsgChan chan string) {
	defer swg.Done()

	executorConfig := execute.ExecutorConfig{
		ICP: map[string]execute.ICPConfig{
			"icpPlus": {APIName: []string{apiName}},
		},
	}
	taskInfo := execute.ExecutorTaskInfo{
		MainTaskInfo: execute.MainTaskInfo{
			WorkspaceId:    "test",
			Target:         "厦门享联科技有限公司",
			OrgId:          "test",
			MainTaskId:     "icp_test",
			ExecutorConfig: executorConfig,
		},
		Executor: "icpPlus",
	}
	taskInfo.MainTaskInfo.ExecutorConfig = executorConfig
	result := icp.Do(taskInfo, true)

	if len(result) > 0 {
		testMsgChan <- fmt.Sprintf("%s-%s: OK!\n", "ICP查询组织备案", apiName)
	} else {
		testMsgChan <- fmt.Sprintf("%s-%s: fail\n", "ICP查询组织备案", "apiName")
	}
	return
}

func testLLMApi(apiName string, swg *sizedwaitgroup.SizedWaitGroup, testMsgChan chan string) {
	defer swg.Done()

	executorConfig := execute.ExecutorConfig{
		LLMAPI: map[string]execute.LLMAPIConfig{
			"kimi":     {},
			"deepseek": {},
			"qwen":     {},
		}}
	taskInfo := execute.ExecutorTaskInfo{
		MainTaskInfo: execute.MainTaskInfo{
			WorkspaceId:    "test",
			Target:         "百度在线网络技术（北京）有限公司",
			OrgId:          "test",
			MainTaskId:     "llmtest",
			ExecutorConfig: executorConfig,
		},
	}

	taskInfo.Executor = apiName
	taskInfo.MainTaskInfo.ExecutorConfig = executorConfig
	result := llmapi.Do(taskInfo)
	if len(result) > 0 {
		testMsgChan <- fmt.Sprintf("%s: OK!\n", apiName)
	} else {
		testMsgChan <- fmt.Sprintf("%s: fail\n", apiName)
	}
	return
}
