package core

import (
	"fmt"
	"github.com/hanc00l/nemo_go/v3/pkg/conf"
	"github.com/hanc00l/nemo_go/v3/pkg/logging"
	"github.com/hanc00l/nemo_go/v3/pkg/resource"
	"github.com/hanc00l/nemo_go/v3/pkg/utils"
	"path/filepath"
)

type RequiredResource struct {
	Category string
	Name     string
}

// CheckRequiredResource 检查并同步资源
// syncExistResource：为true，只同步已存在的资源；false，同步不存在的资源
func CheckRequiredResource(resourceList []RequiredResource, syncExistResource bool) error {
	for _, v := range resourceList {
		var r resource.Resource
		var ok bool
		if r, ok = resource.Resources[v.Category][v.Name]; !ok {
			return fmt.Errorf("资源%s:%s非法", v.Category, v.Name)
		}
		filePath := filepath.Join(conf.GetRootPath(), r.Path, r.Name)
		resourceExistCheck := utils.CheckFileExist(filePath)
		if syncExistResource == true && resourceExistCheck == false {
			continue
		}
		if syncExistResource == false && resourceExistCheck == true {
			continue
		}
		if syncExistResource {
			logging.RuntimeLog.Warningf("资源 %s:%s存在，强制更新...", v.Category, v.Name)
		} else {
			logging.RuntimeLog.Warningf("资源%s:%s不存在，请求资源...", v.Category, v.Name)
		}
		args := RequestResourceArgs{
			Category: v.Category,
			Name:     v.Name,
		}
		var resp ResourceResultArgs
		err := CallXClient("RequestResource", &args, &resp)
		if err != nil {
			return fmt.Errorf("请求资源失败:%s", err.Error())
		}
		rdata := resource.Resource{
			Name:  r.Name,
			Type:  r.Type,
			Path:  resp.Path,
			Bytes: resp.Bytes,
			Hash:  resp.Hash,
		}
		if r.Type == resource.ExecuteFile || r.Type == resource.DataFile || r.Type == resource.ConfigFile {
			err = resource.SaveFileResource(&rdata)
		} else if r.Type == resource.Dir {
			err = resource.SaveDirResource(&rdata)
		} else {
			err = fmt.Errorf("不支持的资源类型:%s", r.Type)
		}
		if err != nil {
			return fmt.Errorf("保存资源%s:%s失败:%s", v.Category, v.Name, err.Error())
		}
	}
	return nil
}
