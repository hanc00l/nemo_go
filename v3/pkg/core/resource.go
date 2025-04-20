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

func CheckRequiredResource(resourceList []RequiredResource) error {
	for _, v := range resourceList {
		var r resource.Resource
		var ok bool
		if r, ok = resource.Resources[v.Category][v.Name]; !ok {
			return fmt.Errorf("resource %s:%s is not valid", v.Category, v.Name)
		}
		filePath := filepath.Join(conf.GetRootPath(), r.Path, r.Name)
		if utils.CheckFileExist(filePath) {
			continue
		}
		logging.RuntimeLog.Warningf("resource %s:%s is not available,requesting...", v.Category, v.Name)
		args := RequestResourceArgs{
			Category: v.Category,
			Name:     v.Name,
		}
		var resp ResourceResultArgs
		err := CallXClient("RequestResource", &args, &resp)
		if err != nil {
			return fmt.Errorf("request resource fail:%s", err.Error())
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
		} else if r.Type == resource.DirAndFile {
			err = resource.SaveDirResource(&rdata)
		} else {
			err = fmt.Errorf("unsupported resource type:%s", r.Type)
		}
		if err != nil {
			return fmt.Errorf("save resource %s:%s fail:%s", v.Category, v.Name, err.Error())
		}
	}
	return nil
}
