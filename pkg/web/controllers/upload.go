package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/hanc00l/nemo_go/pkg/comm"
	"github.com/hanc00l/nemo_go/pkg/conf"
	"github.com/hanc00l/nemo_go/pkg/logging"
	"github.com/hanc00l/nemo_go/pkg/task/onlineapi"
	"github.com/hanc00l/nemo_go/pkg/utils"
	"os"
	"path/filepath"
	"strings"
)

type UploadController struct {
	BaseController
}

const (
	thumbnailWidth = 120
)

// UploadScreenshotAction 上传screenshot文件
func (c *UploadController) UploadScreenshotAction() {
	var rData *StatusResponseData
	defer func() {
		r, _ := json.Marshal(rData)
		c.writeByteContent(comm.EncryptData(r))
	}()

	requestData := comm.DecryptData(c.Ctx.Input.RequestBody)
	var sfi comm.ScreenshotFileInfo
	err := json.Unmarshal(requestData, &sfi)
	if err != nil {
		logging.RuntimeLog.Error(err)
		rData = &StatusResponseData{Status: Fail, Msg: err.Error()}
		return
	}
	// check
	if sfi.Port == 0 || sfi.Domain == "" || sfi.Protocol == "" || len(sfi.Content) == 0 {
		rData = &StatusResponseData{Status: Fail, Msg: "empty upload attribute"}
		return
	}
	if !utils.CheckIPV4(sfi.Domain) && !utils.CheckDomain(sfi.Domain) {
		logging.RuntimeLog.Errorf("invalid domain:%s", sfi.Domain)
		rData = &StatusResponseData{Status: Fail, Msg: "invalid domain"}
		return
	}
	if strings.Contains(sfi.Domain, "..") || strings.Contains(sfi.Domain, "/") {
		logging.RuntimeLog.Errorf("invalid domain:%s", sfi.Domain)
		rData = &StatusResponseData{Status: Fail, Msg: "invalid domain"}
		return
	}
	domainPath := filepath.Join(conf.Nemo.Web.ScreenshotPath, sfi.Domain)
	if !utils.MakePath(conf.Nemo.Web.ScreenshotPath) || !utils.MakePath(domainPath) {
		logging.RuntimeLog.Errorf("check upload path fail:%s", domainPath)
		rData = &StatusResponseData{Status: Fail, Msg: "check upload path fail"}
		return
	}
	//保存文件
	fileName := filepath.Join(domainPath, fmt.Sprintf("%d_%s.png", sfi.Port, sfi.Protocol))
	err = os.WriteFile(fileName, sfi.Content, 0666)
	if err != nil {
		logging.RuntimeLog.Errorf("write file %s fail:%v", fileName, err)
		rData = &StatusResponseData{Status: Fail, Msg: err.Error()}
		return
	}
	//生成缩略图
	fileNameThumbnail := filepath.Join(domainPath, fmt.Sprintf("%d_%s_thumbnail.png", sfi.Port, sfi.Protocol))
	if utils.ReSizePicture(fileName, fileNameThumbnail, thumbnailWidth, 0) {
		rData = &StatusResponseData{Status: Success, Msg: ""}
	} else {
		logging.RuntimeLog.Error("generate thumbnail picature fail")
		rData = &StatusResponseData{Status: Fail, Msg: "generate thumbnail picature fail"}
	}

	return
}

// UploadICPInfoAction 上传ICP备案的查询信息
func (c *UploadController) UploadICPInfoAction() {
	var rData *StatusResponseData
	defer func() {
		r, _ := json.Marshal(rData)
		c.writeByteContent(comm.EncryptData(r))
	}()

	requestData := comm.DecryptData(c.Ctx.Input.RequestBody)
	icpInfos := make(map[string]*onlineapi.ICPInfo)
	err := json.Unmarshal(requestData, &icpInfos)
	if err != nil {
		logging.RuntimeLog.Error(err)
		rData = &StatusResponseData{Status: Fail, Msg: err.Error()}
		return
	}
	if icpInfos == nil || len(icpInfos) <= 0 {
		logging.RuntimeLog.Error(err)
		rData = &StatusResponseData{Status: Fail, Msg: "empty icpinfo"}
		return
	}
	icp := onlineapi.NewICPQuery(onlineapi.ICPQueryConfig{})
	for k, v := range icpInfos {
		icp.ICPMap[k] = v
	}
	if icp.SaveLocalICPInfo() {
		rData = &StatusResponseData{Status: Success, Msg: ""}
		return
	} else {
		logging.RuntimeLog.Error("save icp fail")
		rData = &StatusResponseData{Status: Fail, Msg: "save icp fail"}
		return
	}
}
