package controllers

import (
	"github.com/hanc00l/nemo_go/v2/pkg/conf"
	"github.com/hanc00l/nemo_go/v2/pkg/logging"
	"os"
	"path/filepath"
	"sort"
)

type AccessLogController struct {
	BaseController
}

// IndexAction 显示列表页
func (c *AccessLogController) IndexAction() {
	c.CheckOneAccessRequest(SuperAdmin, true)
	c.Layout = "base.html"
	c.TplName = "accesslog-list.html"
}

// ListAction 列表的数据
func (c *AccessLogController) ListAction() {
	c.CheckOneAccessRequest(SuperAdmin, true)
	defer c.ServeJSON()

	fileNum, err := c.GetInt("file_num", 0)
	if err != nil {
		logging.RuntimeLog.Error(err.Error())
		c.FailedStatus(err.Error())
		return
	}
	logFiles := c.getAccessLogFileList()
	if fileNum < 0 || fileNum >= len(logFiles) {
		c.FailedStatus("日志文件不存在！")
		return
	}
	logFile := logFiles[fileNum]
	logFilePath := filepath.Join(conf.GetRootPath(), "log", logFile)
	logContent, err := os.ReadFile(logFilePath)
	if err != nil {
		c.FailedStatus(err.Error())
		return
	}
	c.SucceededStatus(string(logContent))
}

// AccessLogFilesAction 获取日志文件列表
func (c *AccessLogController) AccessLogFilesAction() {
	defer c.ServeJSON()
	if !c.CheckOneAccessRequest(SuperAdmin, false) {
		c.FailedStatus("当前用户权限不允许！")
		return
	}

	c.Data["json"] = c.getAccessLogFileList()
}

func (c *AccessLogController) getAccessLogFileList() (logs []string) {
	files, _ := filepath.Glob(filepath.Join(conf.GetRootPath(), "log", "access*.log"))
	for _, file := range files {
		_, f := filepath.Split(file)
		logs = append(logs, f)
	}
	sort.Strings(logs)

	return
}
